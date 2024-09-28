//go:build windows

package updates

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"

	"github.com/prometheus-community/windows_exporter/pkg/types"
)

const (
	Name = "updates"

	FlagCacheDuration = "collector.updates.cache-duration"
)

type Config struct {
	cacheDuration time.Duration `yaml:"cache_duration"`
}

var ConfigDefaults = Config{
	cacheDuration: 24 * time.Hour,
}

type Collector struct {
	mu            sync.Mutex
	cacheDuration *time.Duration
	lastScrape    time.Time

	updates availableUpdateCount

	update *prometheus.Desc
}

type update struct {
	status  windowsUpdateStatus
	seventy string
}

type availableUpdateCount map[update]int

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		cacheDuration: &config.cacheDuration,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		cacheDuration: app.Flag(
			FlagCacheDuration,
			"How long should the Windows Update information be cached for.",
		).Default(ConfigDefaults.cacheDuration.String()).Duration(),
	}
	return c
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	c.update = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "update"),
		"Windows Update information",
		[]string{"severity"},
		nil,
	)

	return nil
}

func (c *Collector) GetName() string { return Name }

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var err error

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastScrape) > *c.cacheDuration {
		c.updates, err = c.getUpdates(logger)

		if err != nil {
			logger.Error("failed to collect update status metrics",
				slog.Any("collector", Name),
			)

			return err
		}

		c.lastScrape = time.Now()
	}

	/*
		for _, update := range updates {
			severity := "unknown"
			ch <- prometheus.MustNewConstMetric(
				c.update,
				prometheus.GaugeValue,
				1,
				severity,
			)
		}

	*/
	return nil
}

// S_FALSE is returned by CoInitialize if it was already called on this thread.
const S_FALSE = 0x00000001

type windowsUpdateStatus int

const (
	WUStatusPending = iota
	WUStatusInProgress
	WUStatusCompleted
	WUStatusCompletedWithErrors
	WUStatusFailed
	WUStatusAborted
)

func (c *Collector) getUpdates(logger *slog.Logger) (availableUpdateCount, error) {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != S_FALSE {
			return nil, err
		}
	}

	defer ole.CoUninitialize()

	// Create a new instance of the WMI object
	mus, err := oleutil.CreateObject("Microsoft.Update.Session")
	if err != nil {
		return nil, fmt.Errorf("create Microsoft.Update.Session: %w", err)
	}
	defer mus.Release()

	// Query the IDispatch interface of the object
	musQueryInterface, err := mus.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("IID_IDispatch: %w", err)
	}
	defer musQueryInterface.Release()

	_, err = oleutil.PutProperty(musQueryInterface, "ClientApplicationID", "windows_exporter")
	if err != nil {
		return nil, fmt.Errorf("put ClientApplicationID: %w", err)
	}

	us, err := oleutil.CallMethod(musQueryInterface, "CreateUpdateSearcher")
	defer func(hc *ole.VARIANT) {
		if us != nil {
			_ = us.Clear()
		}
	}(us)

	if err != nil {
		return nil, fmt.Errorf("create update searcher: %w", err)
	}

	ush := us.ToIDispatch()
	defer ush.Release()

	// lets use the fast local-only query to check if WindowsUpdates service is enabled on the host
	hc, err := oleutil.CallMethod(ush, "GetTotalHistoryCount")
	defer func(hc *ole.VARIANT) {
		if hc != nil {
			_ = hc.Clear()
		}
	}(hc)

	if err != nil {
		logger.Error("Windows Updates service is disabled",
			slog.Any("err", err),
		)

		return nil, nil
	}

	usd := us.ToIDispatch()
	defer usd.Release()

	usr, err := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and Type='Software' and IsHidden=0")
	defer func(usr *ole.VARIANT) {
		if usr != nil {
			_ = usr.Clear()
		}
	}(usr)

	if err != nil {
		return nil, fmt.Errorf("search updates: %w", err)
	}

	usrd := usr.ToIDispatch()
	defer usrd.Release()

	upd, err := oleutil.GetProperty(usrd, "Updates")
	defer func(upd *ole.VARIANT) {
		if upd != nil {
			_ = usr.Clear()
		}
	}(upd)

	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}

	updd := upd.ToIDispatch()
	defer updd.Release()

	countUpdd, err := oleutil.GetProperty(updd, "Count")
	defer func(countUpdd *ole.VARIANT) {
		if countUpdd != nil {
			_ = countUpdd.Clear()
		}
	}(countUpdd)

	if err != nil {
		return nil, fmt.Errorf("get count: %w", err)
	}

	availableUpdates := availableUpdateCount{}
	for i := range int(countUpdd.Val) {
		func(i int) {
			// other available properties can be found here:
			// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/aa386114(v=vs.85)

			itemRaw, err := oleutil.GetProperty(updd, "Item", i)
			defer func(itemRaw *ole.VARIANT) {
				if itemRaw != nil {
					_ = itemRaw.Clear()
				}
			}(itemRaw)

			if err != nil {
				logger.Error("failed to fetch Windows Update history item",
					slog.Any("err", err),
				)

				return
			}

			item := itemRaw.ToIDispatch()
			defer item.Release()

			updateType, err := oleutil.GetProperty(item, "Type")
			defer func(updateType *ole.VARIANT) {
				if updateType != nil {
					_ = updateType.Clear()
				}
			}(updateType)

			if err != nil {
				logger.Error("failed to fetch Windows Update history item type",
					slog.Any("err", err),
				)

				return
			}

			severity, err := oleutil.GetProperty(item, "MsrcSeverity")
			defer func(severity *ole.VARIANT) {
				if severity != nil {
					_ = severity.Clear()
				}
			}(severity)

			if err != nil {
				logger.Error("failed to fetch Windows Update history item severity",
					slog.Any("err", err),
				)

				return
			}

			categories, err := oleutil.GetProperty(item, "Categories")
			defer func(categories *ole.VARIANT) {
				if categories != nil {
					_ = categories.Clear()
				}
			}(categories)

			if err != nil {
				logger.Error("failed to fetch Windows Update history item categories",
					slog.Any("err", err),
				)

				return
			}

			_, _, _ = updateType, severity, categories
		}(i)
	}

	return availableUpdates, nil
}
