// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package collector

import (
	"context"
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

	// ctx and deadline are set by [Collection.NewHandlerWithContext] for a handler that serves one scrape.
	// [prometheus.Collector.Collect] has no context parameter, so the handler carries it.
	ctx      context.Context //nolint:containedctx
	deadline time.Time
}

// NewHandler returns a new Handler that implements a [prometheus.Collector] for the given metrics Collection.
// Each call of Collect may take up to maxScrapeDuration, including the time it waits for an earlier scrape to finish.
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

// NewHandlerWithContext returns a Handler for a single scrape, such as one HTTP request.
// The scrape's deadline is maxScrapeDuration from now, so the time Collect waits for an earlier scrape
// to finish counts against it. Collect returns early once ctx is done, without starting collectors
// that have not started yet.
func (c *Collection) NewHandlerWithContext(ctx context.Context, maxScrapeDuration time.Duration, logger *slog.Logger, collectors []string) (*Handler, error) {
	handler, err := c.NewHandler(maxScrapeDuration, logger, collectors)
	if err != nil {
		return nil, err
	}

	handler.ctx = ctx
	handler.deadline = time.Now().Add(maxScrapeDuration)

	return handler, nil
}

func (p *Handler) Describe(_ chan<- *prometheus.Desc) {}

// Collect sends the collected metrics from each of the Collection to
// prometheus.
func (p *Handler) Collect(ch chan<- prometheus.Metric) {
	ctx, deadline := p.ctx, p.deadline
	if ctx == nil {
		ctx, deadline = context.Background(), time.Now().Add(p.maxScrapeDuration)
	}

	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	p.collection.collectAll(ctx, ch, p.logger)
}
