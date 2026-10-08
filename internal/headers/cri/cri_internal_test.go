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
	"encoding/binary"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
	"google.golang.org/protobuf/encoding/protowire"
)

// message builds protobuf messages for test fixtures.
type message []byte

func (m message) string(num protowire.Number, v string) message {
	m = protowire.AppendTag(m, num, protowire.BytesType)

	return protowire.AppendString(m, v)
}

func (m message) message(num protowire.Number, v message) message {
	m = protowire.AppendTag(m, num, protowire.BytesType)

	return protowire.AppendBytes(m, v)
}

func (m message) varint(num protowire.Number, v uint64) message {
	m = protowire.AppendTag(m, num, protowire.VarintType)

	return protowire.AppendVarint(m, v)
}

func (m message) fixed64(num protowire.Number, v uint64) message {
	m = protowire.AppendTag(m, num, protowire.Fixed64Type)

	return protowire.AppendFixed64(m, v)
}

func labels(num protowire.Number, key, value string) message {
	return message{}.message(num, message{}.string(1, key).string(2, value))
}

// A ListContainersResponse as containerd sends it, including fields the client skips.
func containersResponse() message {
	nanoserver := message{}.
		string(1, "abc").
		string(2, "sandbox-1").
		message(3, message{}.string(1, "nanoserver").varint(2, 2)).
		message(4, message{}.string(1, "mcr.microsoft.com/windows/nanoserver")).
		string(5, "sha256:1234").
		varint(6, uint64(ContainerRunning)).
		varint(7, 1700000000000000000)
	nanoserver = append(nanoserver, labels(8, "io.kubernetes.pod.name", "pod")...)
	nanoserver = append(nanoserver, labels(9, "io.kubernetes.container.restartCount", "2")...)

	// Unknown fields of all wire types must be skipped.
	nanoserver = nanoserver.fixed64(100, 42).varint(101, 1).string(102, "unknown")

	return message{}.
		message(1, nanoserver).
		message(1, message{}.string(1, "def").message(3, message{}.string(1, "other")))
}

func sandboxesResponse() message {
	return message{}.
		message(1, message{}.
			string(1, "sandbox-1").
			message(2, message{}.string(1, "pod").string(2, "uid-1").string(3, "default").varint(4, 1)).
			varint(3, uint64(SandboxReady)).
			varint(4, 1700000000000000000)).
		message(1, message{}.
			string(1, "sandbox-2").
			message(2, message{}.string(1, "old").string(3, "kube-system")).
			varint(3, uint64(SandboxNotReady)))
}

func TestDecodeListContainersResponse(t *testing.T) {
	t.Parallel()

	containers, err := decodeListContainersResponse(containersResponse())
	require.NoError(t, err)
	require.Equal(t, []Container{
		{ID: "abc", PodSandboxID: "sandbox-1", Name: "nanoserver", State: ContainerRunning},
		{ID: "def", Name: "other", State: ContainerCreated},
	}, containers)
}

func TestDecodeListPodSandboxResponse(t *testing.T) {
	t.Parallel()

	sandboxes, err := decodeListPodSandboxResponse(sandboxesResponse())
	require.NoError(t, err)
	require.Equal(t, []PodSandbox{
		{ID: "sandbox-1", Name: "pod", UID: "uid-1", Namespace: "default", State: SandboxReady},
		{ID: "sandbox-2", Name: "old", Namespace: "kube-system", State: SandboxNotReady},
	}, sandboxes)
}

func TestDecodeInvalidMessage(t *testing.T) {
	t.Parallel()

	// A truncated container message.
	truncated := containersResponse()
	truncated = truncated[:len(truncated)-3]

	_, err := decodeListContainersResponse(truncated)
	require.Error(t, err)

	// A nested message with an invalid length.
	_, err = decodeListPodSandboxResponse(message{}.message(1, message{0x12, 0x7f}))
	require.Error(t, err)
}

func TestEncodeListRunningContainersRequest(t *testing.T) {
	t.Parallel()

	want := message{}.message(1, message{}.message(2, message{}.varint(1, uint64(ContainerRunning))))
	require.Equal(t, []byte(want), encodeListRunningContainersRequest())
}

// grpcHandler serves a unary gRPC method like grpc-go does.
func grpcHandler(t *testing.T, wantRequest []byte, response []byte) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 || r.Header.Get("Content-Type") != "application/grpc" || r.Header.Get("TE") != "trailers" {
			http.Error(w, "not a gRPC request", http.StatusUnsupportedMediaType)

			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) < 5 || int(binary.BigEndian.Uint32(body[1:5])) != len(body)-5 || string(body[5:]) != string(wantRequest) {
			w.Header().Set("Content-Type", "application/grpc")
			w.Header().Set("Grpc-Status", "3")
			w.Header().Set("Grpc-Message", "unexpected request")

			return
		}

		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status, Grpc-Message")

		frame := make([]byte, 5+len(response))
		binary.BigEndian.PutUint32(frame[1:5], uint32(len(response)))
		copy(frame[5:], response)

		_, _ = w.Write(frame)

		w.Header().Set("Grpc-Status", "0")
	}
}

func startServer(t *testing.T, handler http.Handler) string {
	t.Helper()

	l := listenPipe(t)

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	server := &http.Server{Handler: handler, Protocols: protocols} //nolint:gosec // test server

	go func() { _ = server.Serve(l) }()

	t.Cleanup(func() { _ = server.Close() })

	return "npipe://" + strings.ReplaceAll(l.path, `\`, "/")
}

func TestClient(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.Handle(runtimeService+"Version", grpcHandler(t, nil,
		message{}.string(1, "0.1.0").string(2, "containerd").string(3, "v2.1.4").string(4, "v1")))
	mux.Handle(runtimeService+"ListContainers", grpcHandler(t, encodeListRunningContainersRequest(), containersResponse()))
	mux.Handle(runtimeService+"ListPodSandbox", grpcHandler(t, nil, sandboxesResponse()))

	client, err := NewClient(startServer(t, mux))
	require.NoError(t, err)

	t.Cleanup(client.Close)

	// Repeated calls reuse the HTTP/2 connection.
	for range 3 {
		version, err := client.Version(t.Context())
		require.NoError(t, err)
		require.Equal(t, Version{RuntimeName: "containerd", RuntimeVersion: "v2.1.4", RuntimeAPIVersion: "v1"}, version)

		containers, err := client.ListRunningContainers(t.Context())
		require.NoError(t, err)
		require.Len(t, containers, 2)
		require.Equal(t, "nanoserver", containers[0].Name)

		sandboxes, err := client.ListPodSandboxes(t.Context())
		require.NoError(t, err)
		require.Len(t, sandboxes, 2)
		require.Equal(t, "default", sandboxes[0].Namespace)
	}
}

func TestClientStatusError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc(runtimeService+"Version", func(w http.ResponseWriter, _ *http.Request) {
		// A trailers-only response, as sent for errors.
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Grpc-Status", "12")
		w.Header().Set("Grpc-Message", "unknown service runtime.v1.RuntimeService%2FVersion")
	})

	client, err := NewClient(startServer(t, mux))
	require.NoError(t, err)

	t.Cleanup(client.Close)

	_, err = client.Version(t.Context())

	var statusErr *StatusError

	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, 12, statusErr.Code)
	require.Equal(t, "unknown service runtime.v1.RuntimeService/Version", statusErr.Message)

	// A non-gRPC response, here a 404 for an unregistered method, is rejected.
	_, err = client.ListPodSandboxes(t.Context())
	require.ErrorIs(t, err, ErrInvalidResponse)
}

func TestClientMessageSize(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		payloadSize  int
		declaredSize int
		invalid      bool
	}{
		{name: "at limit", payloadSize: maxMessageSize, declaredSize: maxMessageSize},
		{name: "one byte over limit", payloadSize: maxMessageSize + 1, declaredSize: maxMessageSize + 1, invalid: true},
		{
			name:         "truncated oversized response",
			payloadSize:  maxMessageSize + 2,
			declaredSize: maxMessageSize + 1,
			invalid:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/grpc")
				w.Header().Set("Grpc-Status", "0")

				frame := make([]byte, 5+tc.payloadSize)
				binary.BigEndian.PutUint32(frame[1:5], uint32(tc.declaredSize))

				_, _ = w.Write(frame)
			})

			client, err := NewClient(startServer(t, handler))
			require.NoError(t, err)

			t.Cleanup(client.Close)

			data, err := client.call(t.Context(), "Version", nil)
			if tc.invalid {
				require.ErrorIs(t, err, ErrInvalidResponse)
				require.Nil(t, data)

				return
			}

			require.NoError(t, err)
			require.Len(t, data, tc.payloadSize)
		})
	}
}

func TestClientPipeNotFound(t *testing.T) {
	t.Parallel()

	client, err := NewClient("npipe:////./pipe/windows_exporter-cri-test-does-not-exist")
	require.NoError(t, err)

	t.Cleanup(client.Close)

	_, err = client.Version(t.Context())
	require.ErrorIs(t, err, windows.ERROR_FILE_NOT_FOUND)
}
