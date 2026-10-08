// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package cri

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

// pipePath converts a CRI endpoint like npipe:////./pipe/containerd-containerd
// into a named pipe path like \\.\pipe\containerd-containerd.
func pipePath(endpoint string) (string, error) {
	path, ok := strings.CutPrefix(endpoint, "npipe://")
	if !ok {
		return "", fmt.Errorf("unsupported CRI endpoint %q: only npipe:// endpoints are supported", endpoint)
	}

	path = strings.ReplaceAll(path, "/", `\`)

	if !strings.HasPrefix(path, `\\.\pipe\`) || len(path) == len(`\\.\pipe\`) {
		return "", fmt.Errorf("invalid CRI endpoint %q: expected npipe:////./pipe/<name>", endpoint)
	}

	return path, nil
}

// dialPipe connects to the named pipe. The handle is opened for overlapped I/O,
// so the returned connection supports deadlines and can be closed while a read is pending.
func dialPipe(ctx context.Context, path string) (net.Conn, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	for {
		// SECURITY_ANONYMOUS prevents the pipe server from impersonating the exporter.
		handle, err := windows.CreateFile(
			name,
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_FLAG_OVERLAPPED|windows.SECURITY_SQOS_PRESENT|windows.SECURITY_ANONYMOUS,
			0,
		)
		if err == nil {
			return newReadAheadConn(&pipeConn{File: os.NewFile(uintptr(handle), path)}), nil
		}

		// All pipe instances are busy until the server creates the next one.
		if !errors.Is(err, windows.ERROR_PIPE_BUSY) {
			return nil, &net.OpError{Op: "dial", Net: "pipe", Addr: pipeAddr(path), Err: err}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

type pipeAddr string

func (a pipeAddr) Network() string { return "pipe" }
func (a pipeAddr) String() string  { return string(a) }

// pipeConn adapts a named pipe file to net.Conn.
type pipeConn struct {
	*os.File

	readMu sync.Mutex
	mu     sync.Mutex
	closed bool
}

func (c *pipeConn) LocalAddr() net.Addr  { return pipeAddr(c.Name()) }
func (c *pipeConn) RemoteAddr() net.Addr { return pipeAddr(c.Name()) }

// Read uses event-based overlapped I/O because os.File.Read and os.File.Write
// update a shared file offset concurrently on Windows pipes. Serializing both
// directions would deadlock HTTP/2. Control keeps the handle alive until the
// read completes; File.Close cancels pending pipe I/O, including this read.
func (c *pipeConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if len(p) == 0 {
		return 0, nil
	}

	raw, err := c.SyscallConn()
	if err != nil {
		return 0, err
	}

	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, err
	}

	defer func() { _ = windows.CloseHandle(event) }()

	// The low bit suppresses notifications to the Go runtime's completion port,
	// which only accepts OVERLAPPED structures created by its own poller.
	overlapped := windows.Overlapped{HEvent: event | 1}

	var pinner runtime.Pinner
	pinner.Pin(&overlapped)
	pinner.Pin(&p[0])

	defer pinner.Unpin()

	var (
		n       uint32
		readErr error
	)

	err = raw.Control(func(handle uintptr) {
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()

			readErr = os.ErrClosed

			return
		}

		readErr = windows.ReadFile(windows.Handle(handle), p, &n, &overlapped)
		c.mu.Unlock()

		if errors.Is(readErr, windows.ERROR_IO_PENDING) {
			readErr = windows.GetOverlappedResult(windows.Handle(handle), &overlapped, &n, true)
		}
	})
	if err != nil {
		return 0, err
	}

	switch {
	case errors.Is(readErr, windows.ERROR_BROKEN_PIPE), readErr == nil && n == 0:
		readErr = io.EOF
	case errors.Is(readErr, windows.ERROR_OPERATION_ABORTED):
		readErr = os.ErrClosed
	}

	return int(n), readErr
}

func (c *pipeConn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()

		return os.ErrClosed
	}

	// Prevent a new read from starting after File.Close cancels pending I/O.
	c.closed = true
	c.mu.Unlock()

	return c.File.Close()
}

// Read deadlines are unsupported by the event-based reader. Close cancels it.
func (c *pipeConn) SetReadDeadline(time.Time) error { return nil }

func (c *pipeConn) SetDeadline(t time.Time) error { return c.SetWriteDeadline(t) }

// readAheadConn reads the connection eagerly in the background.
//
// Both HTTP/2 peers write their preface before they read. A write to a named
// pipe without buffer, like the one containerd listens on, blocks until the
// peer reads it, so both sides would block forever on their first write.
type readAheadConn struct {
	net.Conn

	mu   sync.Mutex
	cond *sync.Cond
	buf  []byte
	err  error
}

func newReadAheadConn(conn net.Conn) *readAheadConn {
	c := &readAheadConn{Conn: conn}
	c.cond = sync.NewCond(&c.mu)

	go c.readLoop()

	return c
}

func (c *readAheadConn) readLoop() {
	buf := make([]byte, 32*1024)

	for {
		n, err := c.Conn.Read(buf)

		c.mu.Lock()
		c.buf = append(c.buf, buf[:n]...)
		c.err = err
		c.cond.Broadcast()
		c.mu.Unlock()

		if err != nil {
			return
		}
	}
}

func (c *readAheadConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for len(c.buf) == 0 && c.err == nil {
		c.cond.Wait()
	}

	if len(c.buf) == 0 {
		return 0, c.err
	}

	n := copy(p, c.buf)
	c.buf = c.buf[n:]

	if len(c.buf) == 0 {
		c.buf = nil
	}

	return n, nil
}

// SetDeadline only applies to writes. A read deadline would stop the background reader for good.
func (c *readAheadConn) SetDeadline(t time.Time) error {
	return c.SetWriteDeadline(t)
}

// SetReadDeadline is not supported. Closing the connection unblocks pending reads.
func (c *readAheadConn) SetReadDeadline(time.Time) error {
	return nil
}
