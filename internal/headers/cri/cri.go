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

// Package cri is a minimal client for the Kubernetes Container Runtime Interface (CRI) v1.
//
// gRPC is HTTP/2 with length-prefixed protobuf messages, so the client uses
// net/http with unencrypted HTTP/2 and decodes the few messages it needs by hand.
// This avoids linking google.golang.org/grpc.
// https://github.com/kubernetes/cri-api/blob/master/pkg/apis/runtime/v1/api.proto
package cri

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// maxMessageSize limits the size of a response, like the gRPC default receive limit.
	maxMessageSize = 16 << 20

	runtimeService = "/runtime.v1.RuntimeService/"
)

var ErrInvalidResponse = errors.New("invalid gRPC response")

// StatusError is a non-OK gRPC status returned by the runtime.
type StatusError struct {
	Code    int
	Message string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("gRPC status %d: %s", e.Code, e.Message)
}

// Client calls the CRI runtime service. It keeps one HTTP/2 connection open and is safe for concurrent use.
type Client struct {
	transport *http.Transport
	client    *http.Client
}

// NewClient creates a client for a CRI endpoint like npipe:////./pipe/containerd-containerd.
// It does not connect until the first call.
func NewClient(endpoint string) (*Client, error) {
	path, err := pipePath(endpoint)
	if err != nil {
		return nil, err
	}

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	transport := &http.Transport{
		Protocols: protocols,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialPipe(ctx, path)
		},
		DisableCompression: true,
	}

	return &Client{
		transport: transport,
		client:    &http.Client{Transport: transport},
	}, nil
}

// Close closes the connection to the runtime.
func (c *Client) Close() {
	c.transport.CloseIdleConnections()
}

// Version returns the name and version of the runtime.
func (c *Client) Version(ctx context.Context) (Version, error) {
	resp, err := c.call(ctx, "Version", nil)
	if err != nil {
		return Version{}, err
	}

	return decodeVersion(resp)
}

// ListRunningContainers returns all running containers.
func (c *Client) ListRunningContainers(ctx context.Context) ([]Container, error) {
	resp, err := c.call(ctx, "ListContainers", encodeListRunningContainersRequest())
	if err != nil {
		return nil, err
	}

	return decodeListContainersResponse(resp)
}

// ListPodSandboxes returns all pod sandboxes.
func (c *Client) ListPodSandboxes(ctx context.Context) ([]PodSandbox, error) {
	resp, err := c.call(ctx, "ListPodSandbox", nil)
	if err != nil {
		return nil, err
	}

	return decodeListPodSandboxResponse(resp)
}

// call performs a unary gRPC call and returns the response message.
func (c *Client) call(ctx context.Context, method string, msg []byte) ([]byte, error) {
	// A gRPC message is prefixed by a compression flag and its length.
	body := make([]byte, 5+len(msg))
	binary.BigEndian.PutUint32(body[1:5], uint32(len(msg)))
	copy(body[5:], msg)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost"+runtimeService+method, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("TE", "trailers")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CRI %s: %w", method, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5+maxMessageSize+1))
	if err != nil {
		return nil, fmt.Errorf("CRI %s: %w", method, err)
	}

	if err := grpcStatus(resp); err != nil {
		return nil, fmt.Errorf("CRI %s: %w", method, err)
	}

	if len(data) < 5 || data[0] != 0 {
		return nil, fmt.Errorf("CRI %s: %w: missing or compressed message", method, ErrInvalidResponse)
	}

	if int(binary.BigEndian.Uint32(data[1:5])) != len(data)-5 {
		return nil, fmt.Errorf("CRI %s: %w: message length mismatch", method, ErrInvalidResponse)
	}

	return data[5:], nil
}

// grpcStatus returns the gRPC status of a response, which is sent as trailer.
// Responses without a message carry the status in the headers.
func grpcStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: HTTP status %s", ErrInvalidResponse, resp.Status)
	}

	header := resp.Trailer
	if header.Get("Grpc-Status") == "" {
		header = resp.Header
	}

	status := header.Get("Grpc-Status")
	if status == "" {
		return fmt.Errorf("%w: missing grpc-status", ErrInvalidResponse)
	}

	code, err := strconv.Atoi(status)
	if err != nil {
		return fmt.Errorf("%w: invalid grpc-status %q", ErrInvalidResponse, status)
	}

	if code == 0 {
		return nil
	}

	message := header.Get("Grpc-Message")
	if unescaped, err := url.PathUnescape(message); err == nil {
		message = unescaped
	}

	return &StatusError{Code: code, Message: message}
}
