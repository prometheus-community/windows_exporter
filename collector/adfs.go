// +build windows

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("adfs", newADFSCollector)
}

type adfsCollector struct {
	adLoginConnectionFailures        *prometheus.Desc
	certificateAuthentications       *prometheus.Desc
	deviceAuthentications            *prometheus.Desc
	extranetAccountLockouts          *prometheus.Desc
	federatedAuthentications         *prometheus.Desc
	passportAuthentications          *prometheus.Desc
	passiveRequests                  *prometheus.Desc
	passwordChangeFailed             *prometheus.Desc
	passwordChangeSucceeded          *prometheus.Desc
	tokenRequests                    *prometheus.Desc
	windowsIntegratedAuthentications *prometheus.Desc
}

// newADFSCollector constructs a new adfsCollector
func newADFSCollector() (Collector, error) {
	const subsystem = "adfs"

	return &adfsCollector{
		adLoginConnectionFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ad_login_connection_failures"),
			"Total number of connection failures to an Active Directory domain controller",
			nil,
			nil,
		),
		certificateAuthentications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "certificate_authentications"),
			"Total number of User Certificate authentications",
			nil,
			nil,
		),
		deviceAuthentications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "device_authentications"),
			"Total number of Device authentications",
			nil,
			nil,
		),
		extranetAccountLockouts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "extranet_account_lockouts"),
			"Total number of Extranet Account Lockouts",
			nil,
			nil,
		),
		federatedAuthentications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "federated_authentications"),
			"Total number of authentications from a federated source",
			nil,
			nil,
		),
		passportAuthentications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "passport_authentications"),
			"Total number of Microsoft Passport SSO authentications",
			nil,
			nil,
		),
		passiveRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "passive_requests"),
			"Total number of passive (browser-based) requests",
			nil,
			nil,
		),
		passwordChangeFailed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "password_change_failed"),
			"Total number of failed password changes",
			nil,
			nil,
		),
		passwordChangeSucceeded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "password_change_succeeded"),
			"Total number of successful password changes",
			nil,
			nil,
		),
		tokenRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "token_requests"),
			"Total number of token requests",
			nil,
			nil,
		),
		windowsIntegratedAuthentications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "windows_integrated_authentications"),
			"Total number of Windows integrated authentications (Kerberos/NTLM)",
			nil,
			nil,
		),
	}, nil
}

type perflibADFS struct {
	AdLoginConnectionFailures        float64 `perflib:"AD login Connection Failures"`
	CertificateAuthentications       float64 `perflib:"Certificate Authentications"`
	DeviceAuthentications            float64 `perflib:"Device Authentications"`
	ExtranetAccountLockouts          float64 `perflib:"Extranet Account Lockouts"`
	FederatedAuthentications         float64 `perflib:"Federated Authentications"`
	PassportAuthentications          float64 `perflib:"Microsoft Passport Authentications"`
	PassiveRequests                  float64 `perflib:"Passive Requests"`
	PasswordChangeFailed             float64 `perflib:"Password Change Failed Requests"`
	PasswordChangeSucceeded          float64 `perflib:"Password Change Successful Requests"`
	TokenRequests                    float64 `perflib:"Token Requests"`
	WindowsIntegratedAuthentications float64 `perflib:"Windows Integrated Authentications"`
}

func (c *adfsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var adfsData []perflibADFS
	err := unmarshalObject(ctx.perfObjects["AD FS"], &adfsData)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.adLoginConnectionFailures,
		prometheus.CounterValue,
		adfsData[0].AdLoginConnectionFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.certificateAuthentications,
		prometheus.CounterValue,
		adfsData[0].CertificateAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deviceAuthentications,
		prometheus.CounterValue,
		adfsData[0].DeviceAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.extranetAccountLockouts,
		prometheus.CounterValue,
		adfsData[0].ExtranetAccountLockouts,
	)

	ch <- prometheus.MustNewConstMetric(
		c.federatedAuthentications,
		prometheus.CounterValue,
		adfsData[0].FederatedAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passportAuthentications,
		prometheus.CounterValue,
		adfsData[0].PassportAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passiveRequests,
		prometheus.CounterValue,
		adfsData[0].PassiveRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passwordChangeFailed,
		prometheus.CounterValue,
		adfsData[0].PasswordChangeFailed,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passwordChangeSucceeded,
		prometheus.CounterValue,
		adfsData[0].PasswordChangeSucceeded,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tokenRequests,
		prometheus.CounterValue,
		adfsData[0].TokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.windowsIntegratedAuthentications,
		prometheus.CounterValue,
		adfsData[0].WindowsIntegratedAuthentications,
	)
	return nil
}
