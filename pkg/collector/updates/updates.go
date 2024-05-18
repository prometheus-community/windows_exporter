//go:build windows

package updates

import (
	"errors"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"runtime"
	"time"

	"github.com/prometheus-community/windows_exporter/pkg/types"
)

const (
	Name = "updates"

	FlagUpdatesInclude = "collector.updates.include"
	FlagUpdatesExclude = "collector.updates.exclude"
)

type Config struct {
	printerInclude string `yaml:"printer_include"`
	printerExclude string `yaml:"printer_exclude"`
}

var ConfigDefaults = Config{
	printerInclude: ".+",
	printerExclude: "",
}

type collector struct {
	logger log.Logger

	printerInclude *string
	printerExclude *string

	update *prometheus.Desc

	printerIncludePattern *regexp.Regexp
	printerExcludePattern *regexp.Regexp
}

type update struct {
	title        string
	updateStatus WindowsUpdateStatus
}

type updates []update

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}
	c := &collector{
		printerInclude: &config.printerInclude,
		printerExclude: &config.printerExclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		printerInclude: app.Flag(
			FlagUpdatesInclude,
			"Regular expression to match printers to collect metrics for",
		).Default(ConfigDefaults.printerInclude).String(),
		printerExclude: app.Flag(
			FlagUpdatesExclude,
			"Regular expression to match printers to exclude",
		).Default(ConfigDefaults.printerExclude).String(),
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
		[]string{"severity", "kb"},
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

	for _, update := range updates {
		severity := "unknown"
		ch <- prometheus.MustNewConstMetric(
			c.update,
			prometheus.GaugeValue,
			1,
			severity,
			update.title,
		)
	}

	return nil
}

// S_FALSE is returned by CoInitialize if it was already called on this thread.
const S_FALSE = 0x00000001

type WindowsUpdateStatus int

const (
	WUStatusPending = iota
	WUStatusInProgress
	WUStatusCompleted
	WUStatusCompletedWithErrors
	WUStatusFailed
	WUStatusAborted
)

func (c *collector) getUpdates() (updates, error) {
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
	oleutil.PutProperty(musQueryInterface, "ClientApplicationID", "Cagent")

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

	_, err = oleutil.PutProperty(usd, "Online", "false")

	usr, err := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and Type='Software' and IsHidden=0")
	if err != nil {
		return nil, fmt.Errorf("search for updates: %w", err)
	}

	// Query the IDispatch interface of the object
	usrd := usr.ToIDispatch()
	defer usrd.Release()

	upd, err2 := oleutil.GetProperty(usrd, "Updates")
	if err2 != nil {
		return nil, fmt.Errorf("get updates: %w", err2)
	}

	updd := upd.ToIDispatch()
	defer updd.Release()

	thc, err2 := oleutil.CallMethod(usd, "GetTotalHistoryCount")
	if err2 != nil {
		return nil, fmt.Errorf("get total history count: %w", err2)
	}

	thcn := int(thc.Val)

	// uhistRaw is list of update event records on the computer in descending chronological order
	uhistRaw, err2 := oleutil.CallMethod(usd, "QueryHistory", 0, thcn)
	if err2 != nil {
		return nil, fmt.Errorf("query history: %w", err2)
	}

	uhist := uhistRaw.ToIDispatch()
	defer uhist.Release()

	countUhist, err2 := oleutil.GetProperty(uhist, "Count")
	if err2 != nil {
		return nil, fmt.Errorf("get count: %w", err2)
	}

	availableUpdates := make(updates, 0, countUhist.Val)
	var lastTimeUpdated time.Time

	for i := 0; i < int(countUhist.Val); i++ {
		// other available properties can be found here:
		// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/aa386472(v%3dvs.85)

		itemRaw, err := oleutil.GetProperty(uhist, "Item", i)
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item", "err", err)
			continue
		}

		item := itemRaw.ToIDispatch()
		defer item.Release()

		resultCode, err := oleutil.GetProperty(item, "ResultCode")
		if err != nil {
			// On Win10 machine returns "Exception occurred." after 75 updates so it looks like some undocumented internal limit.
			// We only need the last ones to found "Pending" updates so just ignore this error
			continue
		}

		updateIdentity, err := oleutil.GetProperty(item, "UpdateIdentity")
		if err != nil {
			_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item result code", "err", err)
			continue
		}
		fmt.Println(updateIdentity.ToString())
		fmt.Println(int(resultCode.Val))
		updateStatus := WindowsUpdateStatus(int(resultCode.Val))
		if updateStatus == WUStatusPending || updateStatus == WUStatusInProgress {
			updateIdentity, err := oleutil.GetProperty(item, "Title")
			if err != nil {
				_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item result code", "err", err)
				continue
			}

			availableUpdates = append(availableUpdates, update{
				title: updateIdentity.ToString(),

				updateStatus: updateStatus,
			})
		}

		if updateStatus == WUStatusCompleted {
			date, err := oleutil.GetProperty(item, "Date")
			if err != nil {
				_ = level.Error(c.logger).Log("msg", "failed to fetch Windows Update history item date", "err", err)
				continue
			}
			if updateDate, ok := date.Value().(time.Time); ok {
				if lastTimeUpdated.IsZero() || updateDate.After(lastTimeUpdated) {
					lastTimeUpdated = updateDate
				}
			}
		}
	}

	return availableUpdates, nil
}
