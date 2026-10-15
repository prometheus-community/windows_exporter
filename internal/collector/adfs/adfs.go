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

package adfs

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "adfs"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	adLoginConnectionFailures                          *prometheus.Desc
	artifactDBFailures                                 *prometheus.Desc
	artifactDBQueries                                  *prometheus.Desc
	artifactDBQueryTime                                *prometheus.Desc
	configDBQueries                                    *prometheus.Desc
	configDBQueryTime                                  *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.buildDescriptors()

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "AD FS", nil)
	if err != nil {
		return fmt.Errorf("failed to create AD FS collector: %w", err)
	}

	return nil
}

func (c *Collector) buildDescriptors() {
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
	c.artifactDBQueryTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_artifact_query_time_seconds_total"),
		"Total time spent on artifact database queries, in seconds",
		nil,
		nil,
	)
	c.artifactDBQueries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_artifact_queries_total"),
		"Total number of artifact database queries",
		nil,
		nil,
	)
	c.configDBFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_config_failure_total"),
		"Total number of failures connecting to the configuration database",
		nil,
		nil,
	)
	c.configDBQueryTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_config_query_time_seconds_total"),
		"Total time spent on configuration database queries, in seconds",
		nil,
		nil,
	)
	c.configDBQueries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "db_config_queries_total"),
		"Total number of configuration database queries",
		nil,
		nil,
	)
	c.federationMetadataRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "federation_metadata_requests_total"),
		"Total number of Federation Metadata requests",
		nil,
		nil,
	)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect ADFS metrics: %w", err)
	} else if len(c.perfDataObject) == 0 {
		return fmt.Errorf("failed to collect ADFS metrics: %w", types.ErrNoDataUnexpected)
	}

	// The query times are PERF_AVERAGE_TIMER counters: the raw value is the
	// summed duration in ticks of the counter's time base.
	artifactDBFrequency, _ := c.perfDataCollector.Frequency(counterArtifactDBQueryTime)
	configDBFrequency, _ := c.perfDataCollector.Frequency(counterConfigDBQueryTime)

	c.collectMetrics(ch, c.perfDataObject[0], artifactDBFrequency, configDBFrequency)

	return nil
}

// collectMetrics sends the metrics of data. A query time is only sent if the
// frequency of its counter is known.
func (c *Collector) collectMetrics(ch chan<- prometheus.Metric, data perfDataCounterValues, artifactDBFrequency, configDBFrequency int64) {
	ch <- prometheus.MustNewConstMetric(
		c.adLoginConnectionFailures,
		prometheus.CounterValue,
		data.AdLoginConnectionFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.certificateAuthentications,
		prometheus.CounterValue,
		data.CertificateAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deviceAuthentications,
		prometheus.CounterValue,
		data.DeviceAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.extranetAccountLockouts,
		prometheus.CounterValue,
		data.ExtranetAccountLockouts,
	)

	ch <- prometheus.MustNewConstMetric(
		c.federatedAuthentications,
		prometheus.CounterValue,
		data.FederatedAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passportAuthentications,
		prometheus.CounterValue,
		data.PassportAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passiveRequests,
		prometheus.CounterValue,
		data.PassiveRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passwordChangeFailed,
		prometheus.CounterValue,
		data.PasswordChangeFailed,
	)

	ch <- prometheus.MustNewConstMetric(
		c.passwordChangeSucceeded,
		prometheus.CounterValue,
		data.PasswordChangeSucceeded,
	)

	ch <- prometheus.MustNewConstMetric(
		c.tokenRequests,
		prometheus.CounterValue,
		data.TokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.windowsIntegratedAuthentications,
		prometheus.CounterValue,
		data.WindowsIntegratedAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthAuthZRequests,
		prometheus.CounterValue,
		data.OAuthAuthZRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientAuthentications,
		prometheus.CounterValue,
		data.OAuthClientAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientAuthenticationsFailures,
		prometheus.CounterValue,
		data.OAuthClientAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientCredentialsRequestFailures,
		prometheus.CounterValue,
		data.OAuthClientCredentialRequestFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientCredentialsRequests,
		prometheus.CounterValue,
		data.OAuthClientCredentialRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientPrivateKeyJwtAuthenticationFailures,
		prometheus.CounterValue,
		data.OAuthClientPrivateKeyJWTAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientPrivateKeyJwtAuthentications,
		prometheus.CounterValue,
		data.OAuthClientPrivateKeyJWTAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretBasicAuthenticationFailures,
		prometheus.CounterValue,
		data.OAuthClientBasicAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretBasicAuthentications,
		prometheus.CounterValue,
		data.OAuthClientBasicAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretPostAuthenticationFailures,
		prometheus.CounterValue,
		data.OAuthClientSecretPostAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretPostAuthentications,
		prometheus.CounterValue,
		data.OAuthClientSecretPostAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientWindowsIntegratedAuthenticationFailures,
		prometheus.CounterValue,
		data.OAuthClientWindowsAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientWindowsIntegratedAuthentications,
		prometheus.CounterValue,
		data.OAuthClientWindowsAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthLogonCertificateRequestFailures,
		prometheus.CounterValue,
		data.OAuthLogonCertRequestFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthLogonCertificateTokenRequests,
		prometheus.CounterValue,
		data.OAuthLogonCertTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthPasswordGrantRequestFailures,
		prometheus.CounterValue,
		data.OAuthPasswordGrantRequestFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthPasswordGrantRequests,
		prometheus.CounterValue,
		data.OAuthPasswordGrantRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthTokenRequests,
		prometheus.CounterValue,
		data.OAuthTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samlPTokenRequests,
		prometheus.CounterValue,
		data.SamlPTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ssoAuthenticationFailures,
		prometheus.CounterValue,
		data.SsoAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ssoAuthentications,
		prometheus.CounterValue,
		data.SsoAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wsFedTokenRequests,
		prometheus.CounterValue,
		data.WsFedTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wsTrustTokenRequests,
		prometheus.CounterValue,
		data.WsTrustTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.upAuthenticationFailures,
		prometheus.CounterValue,
		data.UsernamePasswordAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.upAuthentications,
		prometheus.CounterValue,
		data.UsernamePasswordAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.externalAuthenticationFailures,
		prometheus.CounterValue,
		data.ExternalAuthNFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.externalAuthentications,
		prometheus.CounterValue,
		data.ExternalAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.artifactDBFailures,
		prometheus.CounterValue,
		data.ArtifactDBFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.artifactDBQueries,
		prometheus.CounterValue,
		data.ArtifactDBQueries,
	)

	if artifactDBFrequency > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.artifactDBQueryTime,
			prometheus.CounterValue,
			data.ArtifactDBQueryTimeTicks/float64(artifactDBFrequency),
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.configDBFailures,
		prometheus.CounterValue,
		data.ConfigDBFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.configDBQueries,
		prometheus.CounterValue,
		data.ConfigDBQueries,
	)

	if configDBFrequency > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.configDBQueryTime,
			prometheus.CounterValue,
			data.ConfigDBQueryTimeTicks/float64(configDBFrequency),
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.federationMetadataRequests,
		prometheus.CounterValue,
		data.FederationMetadataRequests,
	)
}
