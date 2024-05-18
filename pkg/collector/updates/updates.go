//go:build windows

package updates

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus/client_golang/prometheus"

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

type collector struct {
	logger log.Logger

	cacheDuration *time.Duration
	lastScrape    time.Time

	update *prometheus.Desc
}

type update struct {
	status  windowsUpdateStatus
	seventy string
}

type availableUpdateCount map[update]int

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}
	c := &collector{
		cacheDuration: &config.cacheDuration,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		cacheDuration: app.Flag(
			FlagCacheDuration,
			"How long should the Windows Update information be cached for.",
		).Default(ConfigDefaults.cacheDuration.String()).Duration(),
	}
	return c
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) Build() error {
	c.update = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "update"),
		"Windows Update information",
		[]string{"severity"},
		nil,
	)

	return nil
}

func (c *collector) GetName() string { return Name }

func (c *collector) GetPerfCounter() ([]string, error) { return []string{}, nil }

func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	updates, err := c.getUpdates()

	if err != nil {
		_ = level.Error(c.logger).Log("msg", "failed to collect printer status metrics", "err", err)
		return err
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

	_ = updates

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

func (c *collector) getUpdates() (availableUpdateCount, error) {
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
	if err != nil {
		return nil, fmt.Errorf("create update searcher: %w", err)
	}

	ush := us.ToIDispatch()
	defer ush.Release()
	// lets use the fast local-only query to check if WindowsUpdates service is enabled on the host
	_, err = oleutil.CallMethod(ush, "GetTotalHistoryCount")
	if err != nil {
		_ = level.Error(c.logger).Log("msg", "Windows Updates service is disabled", "err", err)
		return nil, nil
	}

	usd := us.ToIDispatch()
	defer usd.Release()

	usr, err := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and Type='Software' and IsHidden=0")
	if err != nil {
		return nil, fmt.Errorf("search updates: %w", err)
	}

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
		return nil, fmt.Errorf("get count: %w", err)
	}

	availableUpdates := availableUpdateCount{}
	for i := 0; i < int(countUpdd.Val); i++ {
		// other available properties can be found here:
		// https://learn.microsoft.com/en-us/previous-versions/windows/desktop/aa386114(v=vs.85)

		itemRaw, err := oleutil.GetProperty(updd, "Item", i)
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item", "err", err)
			continue
		}

		item := itemRaw.ToIDispatch()
		defer item.Release()

		updateType, err := oleutil.GetProperty(item, "Type")
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item type", "err", err)
			continue
		}

		severity, err := oleutil.GetProperty(item, "MsrcSeverity")
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item severity", "err", err)
			continue
		}

		categories, err := oleutil.GetProperty(item, "Categories")
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item categories", "err", err)
			continue
		}

		_, _, _ = updateType, severity, categories
	}

	return availableUpdates, nil
}
