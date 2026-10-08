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

package update

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/ole"
	"github.com/prometheus-community/windows_exporter/internal/ole/wuapi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "update"

type Config struct {
	Online         bool          `yaml:"online"`
	ScrapeInterval time.Duration `yaml:"scrape-interval"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	Online:         false,
	ScrapeInterval: 6 * time.Hour,
}

var (
	ErrNoUpdates             = errors.New("pending gather update metrics")
	ErrUpdateServiceDisabled = errors.New("windows updates service is disabled")
)

type Collector struct {
	config Config

	mu          sync.RWMutex
	ctxCancelFn context.CancelFunc

	logger *slog.Logger

	metricsBuf []prometheus.Metric

	pendingUpdate              *prometheus.Desc
	pendingUpdateLastPublished *prometheus.Desc
	queryDurationSeconds       *prometheus.Desc
	lastScrapeMetric           *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	app.Flag(
		"collector.update.online",
		"Whether to search for updates online.",
	).Default(strconv.FormatBool(ConfigDefaults.Online)).BoolVar(&c.config.Online)

	app.Flag(
		"collector.update.scrape-interval",
		"Define the interval of scraping Windows Update information.",
	).Default(ConfigDefaults.ScrapeInterval.String()).DurationVar(&c.config.ScrapeInterval)

	return c
}

func (c *Collector) Close() error {
	if c.ctxCancelFn != nil {
		c.ctxCancelFn()
		c.ctxCancelFn = nil
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	c.logger.Info("update collector is in an experimental state! The configuration and metrics may change in future. Please report any issues.")

	c.pendingUpdate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_info"),
		"Expose information for a single pending update item",
		[]string{"id", "revision", "category", "severity", "title"},
		nil,
	)

	c.pendingUpdateLastPublished = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_published_timestamp"),
		"Expose last published timestamp for a single pending update item",
		[]string{"id", "revision"},
		nil,
	)

	c.queryDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "scrape_query_duration_seconds"),
		"Duration of the last scrape query to the Windows Update API",
		nil,
		nil,
	)

	c.lastScrapeMetric = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "scrape_timestamp_seconds"),
		"Timestamp of the last scrape",
		nil,
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())

	// scheduleUpdateStatus sends exactly one value to initErrCh, nil after a successful initialization.
	initErrCh := make(chan error, 1)

	go func() {
		err := recovery.Run(func() error {
			c.scheduleUpdateStatus(ctx, logger, initErrCh, c.config.Online)

			return nil
		})
		if err != nil {
			logger.Error("Windows Update worker stopped",
				slog.Any("err", err),
			)

			// Unblock Build if the panic happened during the initialization.
			select {
			case initErrCh <- err:
			default:
			}
		}
	}()

	c.ctxCancelFn = cancel

	if err := <-initErrCh; err != nil {
		return fmt.Errorf("failed to initialize Windows Update collector: %w", err)
	}

	return nil
}

func (c *Collector) GetName() string { return Name }

func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.metricsBuf == nil {
		return ErrNoUpdates
	}

	for _, m := range c.metricsBuf {
		ch <- m
	}

	return nil
}

func (c *Collector) scheduleUpdateStatus(ctx context.Context, logger *slog.Logger, initErrCh chan<- error, online bool) {
	// COM initialization and every interface call stay on the same OS thread.
	runtime.LockOSThread()

	defer runtime.UnlockOSThread()

	if err := ole.Initialize(); err != nil {
		initErrCh <- err

		return
	}

	defer ole.Uninitialize()

	session, err := wuapi.NewUpdateSession()
	if err != nil {
		initErrCh <- fmt.Errorf("create Microsoft.Update.Session: %w", err)

		return
	}
	defer session.Release()

	if err := session.SetUserLocale(1033); err != nil {
		initErrCh <- fmt.Errorf("set UserLocale: %w", err)

		return
	}

	if err := session.SetClientApplicationID("windows_exporter"); err != nil {
		initErrCh <- fmt.Errorf("set ClientApplicationID: %w", err)

		return
	}

	searcher, err := session.CreateUpdateSearcher()
	if err != nil {
		initErrCh <- fmt.Errorf("create update searcher: %w", err)

		return
	}
	defer searcher.Release()

	if err := searcher.SetOnline(online); err != nil {
		initErrCh <- fmt.Errorf("set Online: %w", err)

		return
	}

	// A local history query checks that the Windows Update service is enabled.
	if _, err := searcher.GetTotalHistoryCount(); err != nil {
		initErrCh <- fmt.Errorf("get update history count: %w", errors.Join(ErrUpdateServiceDisabled, err))

		return
	}

	initErrCh <- nil

	var metricsBuf []prometheus.Metric

	for {
		// A panic in a single search must not stop the worker.
		err = recovery.Run(func() error {
			var err error

			metricsBuf, err = c.fetchUpdates(logger, searcher)

			return err
		})
		if err != nil {
			logger.ErrorContext(ctx, "failed to fetch updates",
				slog.Any("err", err),
			)

			c.mu.Lock()
			c.metricsBuf = nil
			c.mu.Unlock()
		} else {
			c.mu.Lock()
			c.metricsBuf = metricsBuf
			c.mu.Unlock()
		}

		// Failed searches also observe the interval and cancellation.
		select {
		case <-time.After(c.config.ScrapeInterval):
		case <-ctx.Done():
			return
		}
	}
}

func (c *Collector) fetchUpdates(logger *slog.Logger, searcher *wuapi.UpdateSearcher) ([]prometheus.Metric, error) {
	metricsBuf := make([]prometheus.Metric, 0, len(c.metricsBuf)*2+1)

	timeStart := time.Now()

	result, err := searcher.Search("IsInstalled=0 and IsHidden=0")
	if err != nil {
		return nil, fmt.Errorf("search for updates: %w", err)
	}

	defer result.Release()

	logger.Debug(fmt.Sprintf("search for updates took %s", time.Since(timeStart)))

	metricsBuf = append(metricsBuf, prometheus.MustNewConstMetric(
		c.queryDurationSeconds,
		prometheus.GaugeValue,
		time.Since(timeStart).Seconds(),
	))

	updates, err := result.Updates()
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}
	defer updates.Release()

	for item, err := range updates.All() {
		if err != nil {
			return nil, fmt.Errorf("enumerate updates: %w", err)
		}

		update, err := c.getUpdateStatus(item)
		if err != nil {
			logger.Error("failed to fetch Windows Update history item",
				slog.Any("err", err),
			)

			continue
		}

		metricsBuf = append(metricsBuf, prometheus.MustNewConstMetric(
			c.pendingUpdate,
			prometheus.GaugeValue,
			1,
			update.identity,
			update.revision,
			update.category,
			update.severity,
			update.title,
		))

		if update.lastPublished != (time.Time{}) {
			metricsBuf = append(metricsBuf, prometheus.MustNewConstMetric(
				c.pendingUpdateLastPublished,
				prometheus.GaugeValue,
				float64(update.lastPublished.Unix()),
				update.identity,
				update.revision,
			))
		}
	}

	metricsBuf = append(metricsBuf, prometheus.MustNewConstMetric(
		c.lastScrapeMetric,
		prometheus.GaugeValue,
		float64(time.Now().UnixMicro())/1e6,
	))

	return metricsBuf, nil
}

type windowsUpdate struct {
	identity      string
	revision      string
	category      string
	severity      string
	title         string
	lastPublished time.Time
}

// getUpdateStatus retrieves the update status of the given item.
// other available properties can be found here:
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/aa386114(v=vs.85)
func (c *Collector) getUpdateStatus(item *wuapi.Update) (windowsUpdate, error) {
	severity, err := item.MsrcSeverity()
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get MsrcSeverity: %w", err)
	}

	categoryName, err := getUpdateCategory(item)
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get category: %w", err)
	}

	title, err := item.Title()
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get Title: %w", err)
	}

	identity, err := item.Identity()
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get Identity: %w", err)
	}
	defer identity.Release()

	updateID, err := identity.UpdateID()
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get UpdateID: %w", err)
	}

	revision, err := identity.RevisionNumber()
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get RevisionNumber: %w", err)
	}

	lastPublished, err := item.LastDeploymentChangeTime()
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get LastDeploymentChangeTime: %w", err)
	}

	lastPublishedDate, err := lastPublished.Time()
	if err != nil {
		c.logger.Debug("failed to convert LastDeploymentChangeTime",
			slog.String("title", title),
			slog.Any("err", err),
		)

		lastPublishedDate = time.Time{}
	}

	return windowsUpdate{
		identity:      updateID,
		revision:      strconv.FormatInt(int64(revision), 10),
		category:      categoryName,
		severity:      severity,
		title:         title,
		lastPublished: lastPublishedDate,
	}, nil
}

func getUpdateCategory(update *wuapi.Update) (string, error) {
	categories, err := update.Categories()
	if err != nil {
		return "", fmt.Errorf("get Categories: %w", err)
	}
	defer categories.Release()

	var categoryName string

	order := int64(math.MaxInt64)

	for category, err := range categories.All() {
		if err != nil {
			return "", fmt.Errorf("enumerate categories: %w", err)
		}

		name, err := category.Name()
		if err != nil {
			return "", fmt.Errorf("get category Name: %w", err)
		}

		categoryOrder, err := category.Order()
		if err != nil {
			return "", fmt.Errorf("get category Order: %w", err)
		}

		if int64(categoryOrder) < order {
			order = int64(categoryOrder)
			categoryName = name
		}
	}

	return categoryName, nil
}
