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
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// pipeCounter makes pipe names unique, the wall clock is too coarse for parallel tests.
var pipeCounter atomic.Int64 //nolint:gochecknoglobals

// pipeListener accepts connections on a named pipe without buffer, like the one containerd creates.
// Like go-winio, it always keeps one pipe instance waiting, so clients never see a missing pipe.
type pipeListener struct {
	path string

	mu     sync.Mutex
	next   windows.Handle
	closed chan struct{}
}

func listenPipe(t *testing.T) *pipeListener {
	t.Helper()

	l := &pipeListener{
		path:   fmt.Sprintf(`\\.\pipe\windows_exporter-cri-test-%d-%d`, os.Getpid(), pipeCounter.Add(1)),
		closed: make(chan struct{}),
	}

	var err error

	l.next, err = l.createInstance()
	require.NoError(t, err)

	t.Cleanup(func() { _ = l.Close() })

	return l
}

func (l *pipeListener) createInstance() (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(l.path)
	if err != nil {
		return 0, err
	}

	return windows.CreateNamedPipe(
		name,
		windows.PIPE_ACCESS_DUPLEX|windows.FILE_FLAG_OVERLAPPED,
		windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT,
		windows.PIPE_UNLIMITED_INSTANCES,
		0, // no output buffer
		0, // no input buffer
		0,
		nil,
	)
}

func (l *pipeListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	handle := l.next
	l.next = 0
	l.mu.Unlock()

	if handle == 0 {
		return nil, net.ErrClosed
	}

	if err := l.connect(handle); err != nil {
		_ = windows.CloseHandle(handle)

		return nil, err
	}

	next, err := l.createInstance()
	if err != nil {
		_ = windows.CloseHandle(handle)

		return nil, err
	}

	l.mu.Lock()
	select {
	case <-l.closed:
		_ = windows.CloseHandle(next)
	default:
		l.next = next
	}
	l.mu.Unlock()

	return &pipeConn{File: os.NewFile(uintptr(handle), l.path)}, nil
}

// connect waits for a client to connect, or for the listener to be closed.
func (l *pipeListener) connect(handle windows.Handle) error {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = windows.CloseHandle(event)
	}()

	overlapped := windows.Overlapped{HEvent: event}

	err = windows.ConnectNamedPipe(handle, &overlapped)
	if errors.Is(err, windows.ERROR_PIPE_CONNECTED) {
		return nil
	} else if !errors.Is(err, windows.ERROR_IO_PENDING) {
		return err
	}

	for {
		select {
		case <-l.closed:
			_ = windows.CancelIoEx(handle, &overlapped)
			_, _ = windows.WaitForSingleObject(event, windows.INFINITE)

			return net.ErrClosed
		default:
		}

		result, err := windows.WaitForSingleObject(event, 50)
		if err != nil {
			return err
		}

		if result == windows.WAIT_OBJECT_0 {
			return nil
		}
	}
}

func (l *pipeListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	select {
	case <-l.closed:
		return nil
	default:
	}

	close(l.closed)

	if l.next != 0 {
		_ = windows.CloseHandle(l.next)
		l.next = 0
	}

	return nil
}

func (l *pipeListener) Addr() net.Addr { return pipeAddr(l.path) }

func TestPipePath(t *testing.T) {
	t.Parallel()

	for endpoint, want := range map[string]string{
		"npipe:////./pipe/containerd-containerd": `\\.\pipe\containerd-containerd`,
		`npipe://\\.\pipe\containerd-containerd`: `\\.\pipe\containerd-containerd`,
		"npipe:////./pipe/":                      "",
		"npipe://./pipe/containerd-containerd":   "",
		"unix:///run/containerd/containerd.sock": "",
		`\\.\pipe\containerd-containerd`:         "",
	} {
		got, err := pipePath(endpoint)
		if want == "" {
			require.Error(t, err, endpoint)

			continue
		}

		require.NoError(t, err, endpoint)
		require.Equal(t, want, got, endpoint)
	}
}

// Closing a pipe must cancel an event-based read while Control holds the handle.
func TestPipeConnClose(t *testing.T) {
	t.Parallel()

	l := listenPipe(t)
	accepted := make(chan net.Conn, 1)
	acceptErr := make(chan error, 1)

	go func() {
		conn, err := l.Accept()
		if err != nil {
			acceptErr <- err

			return
		}

		accepted <- conn
	}()

	client, err := dialPipe(t.Context(), l.path)
	require.NoError(t, err)

	t.Cleanup(func() { _ = client.Close() })

	var server net.Conn

	select {
	case server = <-accepted:
	case err := <-acceptErr:
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("pipe was not accepted")
	}

	t.Cleanup(func() { _ = server.Close() })

	readErr := make(chan error, 1)

	go func() {
		_, err := server.Read(make([]byte, 1))
		readErr <- err
	}()

	// Sending a byte proves the reader can complete an overlapped operation.
	require.NoError(t, client.SetWriteDeadline(time.Now().Add(5*time.Second)))

	_, err = client.Write([]byte{1})
	require.NoError(t, err)

	select {
	case err := <-readErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("pipe read did not complete")
	}

	go func() {
		_, err := server.Read(make([]byte, 1))
		readErr <- err
	}()

	closed := make(chan error, 1)

	go func() { closed <- server.Close() }()

	select {
	case err := <-readErr:
		require.ErrorIs(t, err, os.ErrClosed)
	case <-time.After(5 * time.Second):
		t.Fatal("pending pipe read was not canceled by Close")
	}

	select {
	case err := <-closed:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not release the pipe handle")
	}

	// A read started after Close must return the same error as a canceled read.
	n, err := server.Read(make([]byte, 1))
	require.Zero(t, n)
	require.ErrorIs(t, err, os.ErrClosed)
}

// Both peers write first. On a pipe without buffer, this blocks both writes
// until the peer reads, unless the client reads ahead.
func TestReadAheadConnPreventsDeadlock(t *testing.T) {
	t.Parallel()

	l := listenPipe(t)

	type result struct {
		conn net.Conn
		err  error
	}

	accepted := make(chan result, 1)

	go func() {
		conn, err := l.Accept()
		accepted <- result{conn, err}
	}()

	var (
		client net.Conn
		err    error
	)

	require.Eventually(t, func() bool {
		client, err = dialPipe(t.Context(), l.path)

		return err == nil
	}, 5*time.Second, 10*time.Millisecond)

	defer client.Close()

	server := <-accepted
	require.NoError(t, server.err)

	defer server.conn.Close()

	done := make(chan error, 2)

	go func() {
		// Server: write the preface, then read the peer's.
		if _, err := server.conn.Write([]byte("server-preface")); err != nil {
			done <- err

			return
		}

		buf := make([]byte, len("client-preface"))

		_, err := io.ReadFull(server.conn, buf)
		done <- err
	}()

	go func() {
		// Client: write the preface, then read the peer's.
		if _, err := client.Write([]byte("client-preface")); err != nil {
			done <- err

			return
		}

		buf := make([]byte, len("server-preface"))
		if _, err := io.ReadFull(client, buf); err != nil {
			done <- err

			return
		}

		if string(buf) != "server-preface" {
			done <- fmt.Errorf("unexpected preface %q", buf)

			return
		}

		done <- nil
	}()

	for range 2 {
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("deadlock: both peers are blocked writing")
		}
	}
}

// Closing the connection must stop the background reader and fail pending reads.
func TestReadAheadConnClose(t *testing.T) {
	t.Parallel()

	client, server := net.Pipe()
	defer server.Close()

	conn := newReadAheadConn(client)

	readErr := make(chan error, 1)

	go func() {
		_, err := conn.Read(make([]byte, 1))
		readErr <- err
	}()

	require.NoError(t, conn.Close())

	select {
	case err := <-readErr:
		require.ErrorIs(t, err, io.ErrClosedPipe)
	case <-time.After(5 * time.Second):
		t.Fatal("pending read was not unblocked by Close")
	}
}

// delayedReadConn keeps a canceled read active until the test releases it.
type delayedReadConn struct {
	net.Conn

	started  chan struct{}
	canceled chan struct{}
	finish   <-chan struct{}
}

func (c *delayedReadConn) Read(p []byte) (int, error) {
	close(c.started)

	n, err := c.Conn.Read(p)
	close(c.canceled)
	<-c.finish

	return n, err
}

func TestReadAheadConnCloseWaitsForReader(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		client, server := net.Pipe()
		defer server.Close()

		finish := make(chan struct{})

		release := sync.OnceFunc(func() { close(finish) })
		defer release()

		reader := &delayedReadConn{
			Conn:     client,
			started:  make(chan struct{}),
			canceled: make(chan struct{}),
			finish:   finish,
		}
		conn := newReadAheadConn(reader)
		<-reader.started

		closed := make(chan error, 1)

		go func() { closed <- conn.Close() }()

		<-reader.canceled
		synctest.Wait()

		select {
		case <-closed:
			t.Fatal("Close returned before the background read finished")
		default:
		}

		release()
		require.NoError(t, <-closed)

		_, err := conn.Read(make([]byte, 1))
		require.ErrorIs(t, err, io.ErrClosedPipe)
	})
}
