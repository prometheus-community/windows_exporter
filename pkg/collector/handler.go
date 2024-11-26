// Copyright 2024 The Prometheus Authors
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

package collector

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Interface guard.
var _ prometheus.Collector = (*Handler)(nil)

// Handler implements [prometheus.Collector] for a set of Windows Collection.
type Handler struct {
	maxScrapeDuration time.Duration
	logger            *slog.Logger
	collection        *Collection
}

// NewHandler returns a new Handler that implements a [prometheus.Collector] for the given metrics Collection.
func (c *Collection) NewHandler(maxScrapeDuration time.Duration, logger *slog.Logger, collectors []string) (*Handler, error) {
	collection := c

	if len(collectors) != 0 {
		var err error

		collection, err = c.WithCollectors(collectors)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler with collectors: %w", err)
		}
	}

	return &Handler{
		maxScrapeDuration: maxScrapeDuration,
		collection:        collection,
		logger:            logger,
	}, nil
}

func (p *Handler) Describe(_ chan<- *prometheus.Desc) {}

// Collect sends the collected metrics from each of the Collection to
// prometheus.
func (p *Handler) Collect(ch chan<- prometheus.Metric) {
	p.collection.collectAll(ch, p.logger, p.maxScrapeDuration)
}
