//go:build windows

package adfs

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"slices"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "adfs"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	adLoginConnectionFailures                          *prometheus.Desc
	artifactDBFailures                                 *prometheus.Desc
	avgArtifactDBQueryTime                             *prometheus.Desc
	avgConfigDBQueryTime                               *prometheus.Desc
	certificateAuthentications                         *prometheus.Desc
	configDBFailures                                   *prometheus.Desc
	deviceAuthentications                              *prometheus.Desc
	externalAuthenticationFailures                     *prometheus.Desc
	externalAuthentications                            *prometheus.Desc
	extranetAccountLockouts                            *prometheus.Desc
	federatedAuthentications                           *prometheus.Desc
	federationMetadataRequests                         *prometheus.Desc
	oAuthAuthZRequests                                 *prometheus.Desc
	oAuthClientAuthentications                         *prometheus.Desc
	oAuthClientAuthenticationsFailures                 *prometheus.Desc
	oAuthClientCredentialsRequestFailures              *prometheus.Desc
	oAuthClientCredentialsRequests                     *prometheus.Desc
	oAuthClientPrivateKeyJwtAuthenticationFailures     *prometheus.Desc
	oAuthClientPrivateKeyJwtAuthentications            *prometheus.Desc
	oAuthClientSecretBasicAuthenticationFailures       *prometheus.Desc
	oAuthClientSecretBasicAuthentications              *prometheus.Desc
	oAuthClientSecretPostAuthenticationFailures        *prometheus.Desc
	oAuthClientSecretPostAuthentications               *prometheus.Desc
	oAuthClientWindowsIntegratedAuthenticationFailures *prometheus.Desc
	oAuthClientWindowsIntegratedAuthentications        *prometheus.Desc
	oAuthLogonCertificateRequestFailures               *prometheus.Desc
	oAuthLogonCertificateTokenRequests                 *prometheus.Desc
	oAuthPasswordGrantRequestFailures                  *prometheus.Desc
	oAuthPasswordGrantRequests                         *prometheus.Desc
	oAuthTokenRequests                                 *prometheus.Desc
	passiveRequests                                    *prometheus.Desc
	passportAuthentications                            *prometheus.Desc
	passwordChangeFailed                               *prometheus.Desc
	passwordChangeSucceeded                            *prometheus.Desc
	samlPTokenRequests                                 *prometheus.Desc
	ssoAuthenticationFailures                          *prometheus.Desc
	ssoAuthentications                                 *prometheus.Desc
	tokenRequests                                      *prometheus.Desc
	upAuthenticationFailures                           *prometheus.Desc
	upAuthentications                                  *prometheus.Desc
	windowsIntegratedAuthentications                   *prometheus.Desc
	wsFedTokenRequests                                 *prometheus.Desc
	wsTrustTokenRequests                               *prometheus.Desc
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

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("AD FS", perfdata.InstanceAll, []string{
		adLoginConnectionFailures,
		certificateAuthentications,
		deviceAuthentications,
		extranetAccountLockouts,
		federatedAuthentications,
		passportAuthentications,
		passiveRequests,
		passwordChangeFailed,
		passwordChangeSucceeded,
		tokenRequests,
		windowsIntegratedAuthentications,
		oAuthAuthZRequests,
		oAuthClientAuthentications,
		oAuthClientAuthenticationFailures,
		oAuthClientCredentialRequestFailures,
		oAuthClientCredentialRequests,
		oAuthClientPrivateKeyJWTAuthenticationFailures,
		oAuthClientPrivateKeyJWTAuthentications,
		oAuthClientBasicAuthenticationFailures,
		oAuthClientBasicAuthentications,
		oAuthClientSecretPostAuthenticationFailures,
		oAuthClientSecretPostAuthentications,
		oAuthClientWindowsAuthenticationFailures,
		oAuthClientWindowsAuthentications,
		oAuthLogonCertRequestFailures,
		oAuthLogonCertTokenRequests,
		oAuthPasswordGrantRequestFailures,
		oAuthPasswordGrantRequests,
		oAuthTokenRequests,
		samlPTokenRequests,
		ssoAuthenticationFailures,
		ssoAuthentications,
		wsFedTokenRequests,
		wsTrustTokenRequests,
		usernamePasswordAuthenticationFailures,
		usernamePasswordAuthentications,
		externalAuthentications,
		externalAuthNFailures,
		artifactDBFailures,
		avgArtifactDBQueryTime,
		configDBFailures,
		avgConfigDBQueryTime,
		federationMetadataRequests,
	})
	if err != nil {
		return fmt.Errorf("failed to create AD FS collector: %w", err)
	}

	c.adLoginConnectionFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ad_login_connection_failures_total"),
		"Total number of connection failures to an Active Directory domain controller",
		nil,
		nil,
	)
	c.certificateAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "certificate_authentications_total"),
		"Total number of User Certificate authentications",
		nil,
		nil,
	)
	c.deviceAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "device_authentications_total"),
		"Total number of Device authentications",
		nil,
		nil,
	)
	c.extranetAccountLockouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "extranet_account_lockouts_total"),
		"Total number of Extranet Account Lockouts",
		nil,
		nil,
	)
	c.federatedAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "federated_authentications_total"),
		"Total number of authentications from a federated source",
		nil,
		nil,
	)
	c.passportAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "passport_authentications_total"),
		"Total number of Microsoft Passport SSO authentications",
		nil,
		nil,
	)
	c.passiveRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "passive_requests_total"),
		"Total number of passive (browser-based) requests",
		nil,
		nil,
	)
	c.passwordChangeFailed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "password_change_failed_total"),
		"Total number of failed password changes",
		nil,
		nil,
	)
	c.passwordChangeSucceeded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "password_change_succeeded_total"),
		"Total number of successful password changes",
		nil,
		nil,
	)
	c.tokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "token_requests_total"),
		"Total number of token requests",
		nil,
		nil,
	)
	c.windowsIntegratedAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "windows_integrated_authentications_total"),
		"Total number of Windows integrated authentications (Kerberos/NTLM)",
		nil,
		nil,
	)
	c.oAuthAuthZRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_authorization_requests_total"),
		"Total number of incoming requests to the OAuth Authorization endpoint",
		nil,
		nil,
	)
	c.oAuthClientAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_authentication_success_total"),
		"Total number of successful OAuth client Authentications",
		nil,
		nil,
	)
	c.oAuthClientAuthenticationsFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_authentication_failure_total"),
		"Total number of failed OAuth client Authentications",
		nil,
		nil,
	)
	c.oAuthClientCredentialsRequestFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_credentials_failure_total"),
		"Total number of failed OAuth Client Credentials Requests",
		nil,
		nil,
	)
	c.oAuthClientCredentialsRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_credentials_success_total"),
		"Total number of successful RP tokens issued for OAuth Client Credentials Requests",
		nil,
		nil,
	)
	c.oAuthClientPrivateKeyJwtAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_privkey_jwt_authentication_failure_total"),
		"Total number of failed OAuth Client Private Key Jwt Authentications",
		nil,
		nil,
	)
	c.oAuthClientPrivateKeyJwtAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_privkey_jwt_authentications_success_total"),
		"Total number of successful OAuth Client Private Key Jwt Authentications",
		nil,
		nil,
	)
	c.oAuthClientSecretBasicAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_secret_basic_authentications_failure_total"),
		"Total number of failed OAuth Client Secret Basic Authentications",
		nil,
		nil,
	)
	c.oAuthClientSecretBasicAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_secret_basic_authentications_success_total"),
		"Total number of successful OAuth Client Secret Basic Authentications",
		nil,
		nil,
	)
	c.oAuthClientSecretPostAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_secret_post_authentications_failure_total"),
		"Total number of failed OAuth Client Secret Post Authentications",
		nil,
		nil,
	)
	c.oAuthClientSecretPostAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_secret_post_authentications_success_total"),
		"Total number of successful OAuth Client Secret Post Authentications",
		nil,
		nil,
	)
	c.oAuthClientWindowsIntegratedAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_windows_authentications_failure_total"),
		"Total number of failed OAuth Client Windows Integrated Authentications",
		nil,
		nil,
	)
	c.oAuthClientWindowsIntegratedAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_client_windows_authentications_success_total"),
		"Total number of successful OAuth Client Windows Integrated Authentications",
		nil,
		nil,
	)
	c.oAuthLogonCertificateRequestFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_logon_certificate_requests_failure_total"),
		"Total number of failed OAuth Logon Certificate Requests",
		nil,
		nil,
	)
	c.oAuthLogonCertificateTokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_logon_certificate_token_requests_success_total"),
		"Total number of successful RP tokens issued for OAuth Logon Certificate Requests",
		nil,
		nil,
	)
	c.oAuthPasswordGrantRequestFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_password_grant_requests_failure_total"),
		"Total number of failed OAuth Password Grant Requests",
		nil,
		nil,
	)
	c.oAuthPasswordGrantRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_password_grant_requests_success_total"),
		"Total number of successful OAuth Password Grant Requests",
		nil,
		nil,
	)
	c.oAuthTokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "oauth_token_requests_success_total"),
		"Total number of successful RP tokens issued over OAuth protocol",
		nil,
		nil,
	)
	c.samlPTokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "samlp_token_requests_success_total"),
		"Total number of successful RP tokens issued over SAML-P protocol",
		nil,
		nil,
	)
	c.ssoAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sso_authentications_failure_total"),
		"Total number of failed SSO authentications",
		nil,
		nil,
	)
	c.ssoAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sso_authentications_success_total"),
		"Total number of successful SSO authentications",
		nil,
		nil,
	)
	c.wsFedTokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wsfed_token_requests_success_total"),
		"Total number of successful RP tokens issued over WS-Fed protocol",
		nil,
		nil,
	)
	c.wsTrustTokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wstrust_token_requests_success_total"),
		"Total number of successful RP tokens issued over WS-Trust protocol",
		nil,
		nil,
	)
	c.upAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "userpassword_authentications_failure_total"),
		"Total number of failed AD U/P authentications",
		nil,
		nil,
	)
	c.upAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "userpassword_authentications_success_total"),
		"Total number of successful AD U/P authentications",
		nil,
		nil,
	)
	c.externalAuthenticationFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "external_authentications_failure_total"),
		"Total number of failed authentications from external MFA providers",
		nil,
		nil,
	)
	c.externalAuthentications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "external_authentications_success_total"),
		"Total number of successful authentications from external MFA providers",
		nil,
		nil,
	)
	c.artifactDBFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_artifact_failure_total"),
		"Total number of failures connecting to the artifact database",
		nil,
		nil,
	)
	c.avgArtifactDBQueryTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_artifact_query_time_seconds_total"),
		"Accumulator of time taken for an artifact database query",
		nil,
		nil,
	)
	c.configDBFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_config_failure_total"),
		"Total number of failures connecting to the configuration database",
		nil,
		nil,
	)
	c.avgConfigDBQueryTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_config_query_time_seconds_total"),
		"Accumulator of time taken for a configuration database query",
		nil,
		nil,
	)
	c.federationMetadataRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "federation_metadata_requests_total"),
		"Total number of Federation Metadata requests",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect ADFS metrics: %w", err)
	}

	instanceKey := slices.Collect(maps.Keys(data))

	if len(instanceKey) == 0 {
		return errors.New("perflib query for ADFS returned empty result set")
	}

	adfsData, ok := data[instanceKey[0]]

	if !ok {
		return errors.New("perflib query for ADFS returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.adLoginConnectionFailures,
		prometheus.CounterValue,
		adfsData[adLoginConnectionFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.certificateAuthentications,
		prometheus.CounterValue,
		adfsData[certificateAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deviceAuthentications,
		prometheus.CounterValue,
		adfsData[deviceAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.extranetAccountLockouts,
		prometheus.CounterValue,
		adfsData[extranetAccountLockouts].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.federatedAuthentications,
		prometheus.CounterValue,
		adfsData[federatedAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passportAuthentications,
		prometheus.CounterValue,
		adfsData[passportAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passiveRequests,
		prometheus.CounterValue,
		adfsData[passiveRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passwordChangeFailed,
		prometheus.CounterValue,
		adfsData[passwordChangeFailed].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passwordChangeSucceeded,
		prometheus.CounterValue,
		adfsData[passwordChangeSucceeded].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tokenRequests,
		prometheus.CounterValue,
		adfsData[tokenRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.windowsIntegratedAuthentications,
		prometheus.CounterValue,
		adfsData[windowsIntegratedAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthAuthZRequests,
		prometheus.CounterValue,
		adfsData[oAuthAuthZRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientAuthentications,
		prometheus.CounterValue,
		adfsData[oAuthClientAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientAuthenticationsFailures,
		prometheus.CounterValue,
		adfsData[oAuthClientAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientCredentialsRequestFailures,
		prometheus.CounterValue,
		adfsData[oAuthClientCredentialRequestFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientCredentialsRequests,
		prometheus.CounterValue,
		adfsData[oAuthClientCredentialRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientPrivateKeyJwtAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[oAuthClientPrivateKeyJWTAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientPrivateKeyJwtAuthentications,
		prometheus.CounterValue,
		adfsData[oAuthClientPrivateKeyJWTAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretBasicAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[oAuthClientBasicAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretBasicAuthentications,
		prometheus.CounterValue,
		adfsData[oAuthClientBasicAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretPostAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[oAuthClientSecretPostAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretPostAuthentications,
		prometheus.CounterValue,
		adfsData[oAuthClientSecretPostAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientWindowsIntegratedAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[oAuthClientWindowsAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientWindowsIntegratedAuthentications,
		prometheus.CounterValue,
		adfsData[oAuthClientWindowsAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthLogonCertificateRequestFailures,
		prometheus.CounterValue,
		adfsData[oAuthLogonCertRequestFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthLogonCertificateTokenRequests,
		prometheus.CounterValue,
		adfsData[oAuthLogonCertTokenRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthPasswordGrantRequestFailures,
		prometheus.CounterValue,
		adfsData[oAuthPasswordGrantRequestFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthPasswordGrantRequests,
		prometheus.CounterValue,
		adfsData[oAuthPasswordGrantRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthTokenRequests,
		prometheus.CounterValue,
		adfsData[oAuthTokenRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samlPTokenRequests,
		prometheus.CounterValue,
		adfsData[samlPTokenRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ssoAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[ssoAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ssoAuthentications,
		prometheus.CounterValue,
		adfsData[ssoAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wsFedTokenRequests,
		prometheus.CounterValue,
		adfsData[wsFedTokenRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wsTrustTokenRequests,
		prometheus.CounterValue,
		adfsData[wsTrustTokenRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.upAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[usernamePasswordAuthenticationFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.upAuthentications,
		prometheus.CounterValue,
		adfsData[usernamePasswordAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.externalAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[externalAuthNFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.externalAuthentications,
		prometheus.CounterValue,
		adfsData[externalAuthentications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.artifactDBFailures,
		prometheus.CounterValue,
		adfsData[artifactDBFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.avgArtifactDBQueryTime,
		prometheus.CounterValue,
		adfsData[avgArtifactDBQueryTime].FirstValue*math.Pow(10, -8),
	)

	ch <- prometheus.MustNewConstMetric(
		c.configDBFailures,
		prometheus.CounterValue,
		adfsData[configDBFailures].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.avgConfigDBQueryTime,
		prometheus.CounterValue,
		adfsData[avgConfigDBQueryTime].FirstValue*math.Pow(10, -8),
	)

	ch <- prometheus.MustNewConstMetric(
		c.federationMetadataRequests,
		prometheus.CounterValue,
		adfsData[federationMetadataRequests].FirstValue,
	)

	return nil
}
