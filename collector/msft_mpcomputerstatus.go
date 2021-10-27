package collector

import (
	"github.com/StackExchange/wmi"
	// "github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("msft_mpcomputerstatus", newMSFT_MpComputerStatusCollector) // TODO: Add any perflib dependencies here
}

// A MSFT_MpComputerStatusCollector is a Prometheus collector for WMI MSFT_MpComputerStatus metrics
type MSFT_MpComputerStatusCollector struct {
	AntispywareSignatureAge *prometheus.Desc
	AntivirusSignatureAge   *prometheus.Desc
	ComputerState           *prometheus.Desc
	FullScanAge             *prometheus.Desc
	NISSignatureAge         *prometheus.Desc
	QuickScanAge            *prometheus.Desc
}

func newMSFT_MpComputerStatusCollector() (Collector, error) {
	const subsystem = "msft_mpcomputerstatus"
	return &MSFT_MpComputerStatusCollector{
		AntispywareSignatureAge: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "antispyware_signature_age"),
			"(AntispywareSignatureAge)",
			nil,
			nil,
		),
		AntivirusSignatureAge: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "antivirus_signature_age"),
			"(AntivirusSignatureAge)",
			nil,
			nil,
		),
		ComputerState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "computer_state"),
			"(ComputerState)",
			nil,
			nil,
		),
		FullScanAge: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "full_scan_age"),
			"(FullScanAge)",
			nil,
			nil,
		),
		NISSignatureAge: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "nis_signature_age"),
			"(NISSignatureAge)",
			nil,
			nil,
		),
		QuickScanAge: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quick_scan_age"),
			"(QuickScanAge)",
			nil,
			nil,
		),
	}, nil
}

// MSFT_MpComputerStatus docs:
// - <add link to documentation here>
type MSFT_MpComputerStatus struct {
	Name string

	AntispywareSignatureAge uint32
	AntivirusSignatureAge   uint32
	ComputerState           uint32
	FullScanAge             uint32
	NISSignatureAge         uint32
	QuickScanAge            uint32
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSFT_MpComputerStatusCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSFT_MpComputerStatus
	q := queryAll(&dst)
  
  if err := wmi.QueryNamespace(q, &dst, "root/microsoft/windows/defender"); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AntispywareSignatureAge,
		prometheus.GaugeValue,
		float64(dst[0].AntispywareSignatureAge),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AntivirusSignatureAge,
		prometheus.GaugeValue,
		float64(dst[0].AntivirusSignatureAge),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ComputerState,
		prometheus.GaugeValue,
		float64(dst[0].ComputerState),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FullScanAge,
		prometheus.GaugeValue,
		float64(dst[0].FullScanAge),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NISSignatureAge,
		prometheus.GaugeValue,
		float64(dst[0].NISSignatureAge),
	)

	ch <- prometheus.MustNewConstMetric(
		c.QuickScanAge,
		prometheus.GaugeValue,
		float64(dst[0].QuickScanAge),
	)

	return nil
}
