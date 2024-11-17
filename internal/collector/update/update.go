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
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "update"

type Config struct {
	online         bool          `yaml:"online"`
	scrapeInterval time.Duration `yaml:"scrape_interval"`
}

var ConfigDefaults = Config{
	online:         false,
	scrapeInterval: 6 * time.Hour,
}

var ErrNoUpdates = errors.New("no updates available")

type Collector struct {
	config Config

	mu          sync.RWMutex
	ctxCancelFn context.CancelFunc

	metricsBuf []prometheus.Metric

	pendingUpdate        *prometheus.Desc
	queryDurationSeconds *prometheus.Desc
	lastScrapeMetric     *prometheus.Desc
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
		"collector.updates.online",
		"Whether to search for updates online.",
	).Default(strconv.FormatBool(ConfigDefaults.online)).BoolVar(&c.config.online)

	app.Flag(
		"collector.updates.scrape-interval",
		"Define the interval of scraping Windows Update information.",
	).Default(ConfigDefaults.scrapeInterval.String()).DurationVar(&c.config.scrapeInterval)

	return c
}

func (c *Collector) Close() error {
	c.ctxCancelFn()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	logger = logger.With(slog.String("collector", Name))

	logger.Info("update collector is in an experimental state! The configuration and metrics may change in future. Please report any issues.")

	ctx, cancel := context.WithCancel(context.Background())

	initErrCh := make(chan error, 1)
	go c.scheduleUpdateStatus(ctx, logger, initErrCh, c.config.online)

	c.ctxCancelFn = cancel

	if err := <-initErrCh; err != nil {
		return fmt.Errorf("failed to initialize Windows Update collector: %w", err)
	}

	c.pendingUpdate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pending_info"),
		"Pending Windows Updates",
		[]string{"category", "severity", "title"},
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

	return nil
}

func (c *Collector) GetName() string { return Name }

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
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
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != 0x00000001 {
			initErrCh <- fmt.Errorf("CoInitializeEx: %w", err)

			return
		}
	}

	defer ole.CoUninitialize()

	// Create a new instance of the WMI object
	mus, err := oleutil.CreateObject("Microsoft.Update.Session")
	if err != nil {
		initErrCh <- fmt.Errorf("create Microsoft.Update.Session: %w", err)

		return
	}

	defer mus.Release()

	// Query the IDispatch interface of the object
	musQueryInterface, err := mus.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		initErrCh <- fmt.Errorf("IID_IDispatch: %w", err)

		return
	}

	defer musQueryInterface.Release()

	_, err = oleutil.PutProperty(musQueryInterface, "ClientApplicationID", "windows_exporter")
	if err != nil {
		initErrCh <- fmt.Errorf("put ClientApplicationID: %w", err)

		return
	}

	// https://learn.microsoft.com/en-us/windows/win32/api/wuapi/nf-wuapi-iupdatesession-createupdatesearcher
	us, err := oleutil.CallMethod(musQueryInterface, "CreateUpdateSearcher")
	defer func(hc *ole.VARIANT) {
		if us != nil {
			_ = hc.Clear()
		}
	}(us)

	if err != nil {
		initErrCh <- fmt.Errorf("create update searcher: %w", err)

		return
	}

	ush := us.ToIDispatch()
	defer ush.Release()

	_, err = oleutil.PutProperty(ush, "Online", online)
	if err != nil {
		initErrCh <- fmt.Errorf("put Online: %w", err)

		return
	}

	// lets use the fast local-only query to check if WindowsUpdates service is enabled on the host
	hc, err := oleutil.CallMethod(ush, "GetTotalHistoryCount")
	defer func(hc *ole.VARIANT) {
		if hc != nil {
			_ = hc.Clear()
		}
	}(hc)

	if err != nil {
		initErrCh <- fmt.Errorf("windows updates service is disabled: %w", err)

		return
	}

	close(initErrCh)

	usd := us.ToIDispatch()
	defer usd.Release()

	var metricsBuf []prometheus.Metric

	for {
		metricsBuf, err = c.fetchUpdates(logger, usd)
		if err != nil {
			logger.ErrorContext(ctx, "failed to fetch updates",
				slog.Any("err", err),
			)

			c.mu.Lock()
			c.metricsBuf = nil
			c.mu.Unlock()

			continue
		}

		c.mu.Lock()
		c.metricsBuf = metricsBuf
		c.mu.Unlock()

		select {
		case <-time.After(c.config.scrapeInterval):
		case <-ctx.Done():
			return
		}
	}
}

func (c *Collector) fetchUpdates(logger *slog.Logger, usd *ole.IDispatch) ([]prometheus.Metric, error) {
	metricsBuf := make([]prometheus.Metric, 0, len(c.metricsBuf))

	timeStart := time.Now()

	usr, err := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and IsHidden=0")
	if err != nil {
		return nil, fmt.Errorf("search for updates: %w", err)
	}

	logger.Debug(fmt.Sprintf("search for updates took %s", time.Since(timeStart)))

	metricsBuf = append(metricsBuf, prometheus.MustNewConstMetric(
		c.queryDurationSeconds,
		prometheus.GaugeValue,
		time.Since(timeStart).Seconds(),
	))

	usrd := usr.ToIDispatch()
	defer usrd.Release()

	upd, err := oleutil.GetProperty(usrd, "Updates")
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}

	updd := upd.ToIDispatch()
	defer updd.Release()

	countUpdd, err := oleutil.GetProperty(updd, "Count")
	if err != nil {
		return nil, fmt.Errorf("get updates count: %w", err)
	}

	for i := range int(countUpdd.Val) {
		update, err := c.getUpdateStatus(updd, i)
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
			update.category,
			update.severity,
			update.title,
		))
	}

	metricsBuf = append(metricsBuf, prometheus.MustNewConstMetric(
		c.lastScrapeMetric,
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
	))

	return metricsBuf, nil
}

type windowsUpdate struct {
	category string
	severity string
	title    string
}

// getUpdateStatus retrieves the update status of the given item.
// other available properties can be found here:
// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/aa386114(v=vs.85)
func (c *Collector) getUpdateStatus(updd *ole.IDispatch, item int) (windowsUpdate, error) {
	itemRaw, err := oleutil.GetProperty(updd, "Item", item)
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get update item: %w", err)
	}

	updateItem := itemRaw.ToIDispatch()
	defer updateItem.Release()

	severity, err := oleutil.GetProperty(updateItem, "MsrcSeverity")
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get MsrcSeverity: %w", err)
	}

	categoriesRaw, err := oleutil.GetProperty(updateItem, "Categories")
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get Categories: %w", err)
	}

	categories := categoriesRaw.ToIDispatch()
	defer categories.Release()

	categoryName, err := getUpdateCategory(categories)
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get Category: %w", err)
	}

	title, err := oleutil.GetProperty(updateItem, "Title")
	if err != nil {
		return windowsUpdate{}, fmt.Errorf("get Title: %w", err)
	}

	return windowsUpdate{
		category: categoryName,
		severity: severity.ToString(),
		title:    title.ToString(),
	}, nil
}

func getUpdateCategory(categories *ole.IDispatch) (string, error) {
	var categoryName string

	categoryCount, err := oleutil.GetProperty(categories, "Count")
	if err != nil {
		return categoryName, fmt.Errorf("get Categories count: %w", err)
	}

	order := int64(math.MaxInt64)

	for i := range categoryCount.Val {
		err = func(i int64) error {
			categoryRaw, err := oleutil.GetProperty(categories, "Item", i)
			if err != nil {
				return fmt.Errorf("get Category item: %w", err)
			}

			category := categoryRaw.ToIDispatch()
			defer category.Release()

			categoryNameRaw, err := oleutil.GetProperty(category, "Name")
			if err != nil {
				return fmt.Errorf("get Category item Name: %w", err)
			}

			orderRaw, err := oleutil.GetProperty(category, "Order")
			if err != nil {
				return fmt.Errorf("get Category item Order: %w", err)
			}

			if orderRaw.Val < order {
				order = orderRaw.Val
				categoryName = categoryNameRaw.ToString()
			}

			return nil
		}(i)
		if err != nil {
			return "", fmt.Errorf("get Category item: %w", err)
		}
	}

	return categoryName, nil
}
