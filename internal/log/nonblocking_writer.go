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

package log

import (
	"bytes"
	"io"
	"sync"
)

const defaultNonBlockingWriterBufferSize = 1024

type nonBlockingWriter struct {
	writer     io.Writer
	queue      chan []byte
	stop       chan struct{}
	workerDone chan struct{}
	once       sync.Once
}

func newNonBlockingWriter(writer io.Writer, queueSize int) *nonBlockingWriter {
	if queueSize <= 0 {
		queueSize = defaultNonBlockingWriterBufferSize
	}

	w := &nonBlockingWriter{
		writer:     writer,
		queue:      make(chan []byte, queueSize),
		stop:       make(chan struct{}),
		workerDone: make(chan struct{}),
	}

	go func() {
		defer close(w.workerDone)

		for {
			select {
			case p := <-w.queue:
				_, _ = w.writer.Write(p)
			case <-w.stop:
				for {
					select {
					case p := <-w.queue:
						_, _ = w.writer.Write(p)
					default:
						return
					}
				}
			}
		}
	}()

	return w
}

func (w *nonBlockingWriter) Write(p []byte) (int, error) {
	msg := bytes.Clone(p)

	select {
	case <-w.stop:
		return 0, io.ErrClosedPipe
	case w.queue <- msg:
	default:
	}

	return len(p), nil
}

func (w *nonBlockingWriter) Close() error {
	w.once.Do(func() {
		close(w.stop)
	})

	<-w.workerDone

	return nil
}
