//go:build windows

package adfs

import (
	"log/slog"
	"math"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "adfs"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

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
	wsfedTokenRequests                                 *prometheus.Desc
	wstrustTokenRequests                               *prometheus.Desc
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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{"AD FS"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
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
	c.wsfedTokenRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wsfed_token_requests_success_total"),
		"Total number of successful RP tokens issued over WS-Fed protocol",
		nil,
		nil,
	)
	c.wstrustTokenRequests = prometheus.NewDesc(
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

type perflibADFS struct {
	AdLoginConnectionFailures            float64 `perflib:"AD Login Connection Failures"`
	CertificateAuthentications           float64 `perflib:"Certificate Authentications"`
	DeviceAuthentications                float64 `perflib:"Device Authentications"`
	ExtranetAccountLockouts              float64 `perflib:"Extranet Account Lockouts"`
	FederatedAuthentications             float64 `perflib:"Federated Authentications"`
	PassportAuthentications              float64 `perflib:"Microsoft Passport Authentications"`
	PassiveRequests                      float64 `perflib:"Passive Requests"`
	PasswordChangeFailed                 float64 `perflib:"Password Change Failed Requests"`
	PasswordChangeSucceeded              float64 `perflib:"Password Change Successful Requests"`
	TokenRequests                        float64 `perflib:"Token Requests"`
	WindowsIntegratedAuthentications     float64 `perflib:"Windows Integrated Authentications"`
	OAuthAuthZRequests                   float64 `perflib:"OAuth AuthZ Requests"`
	OAuthClientAuthentications           float64 `perflib:"OAuth Client Authentications"`
	OAuthClientAuthenticationFailures    float64 `perflib:"OAuth Client Authentications Failures"`
	OAuthClientCredentialRequestFailures float64 `perflib:"OAuth Client Credentials Request Failures"`
	OAuthClientCredentialRequests        float64 `perflib:"OAuth Client Credentials Requests"`
	OAuthClientPrivKeyJWTAuthnFailures   float64 `perflib:"OAuth Client Private Key Jwt Authentication Failures"`
	OAuthClientPrivKeyJWTAuthentications float64 `perflib:"OAuth Client Private Key Jwt Authentications"`
	OAuthClientBasicAuthnFailures        float64 `perflib:"OAuth Client Secret Basic Authentication Failures"`
	OAuthClientBasicAuthentications      float64 `perflib:"OAuth Client Secret Basic Authentication Requests"`
	OAuthClientSecretPostAuthnFailures   float64 `perflib:"OAuth Client Secret Post Authentication Failures"`
	OAuthClientSecretPostAuthentications float64 `perflib:"OAuth Client Secret Post Authentications"`
	OAuthClientWindowsAuthnFailures      float64 `perflib:"OAuth Client Windows Integrated Authentication Failures"`
	OAuthClientWindowsAuthentications    float64 `perflib:"OAuth Client Windows Integrated Authentications"`
	OAuthLogonCertRequestFailures        float64 `perflib:"OAuth Logon Certificate Request Failures"`
	OAuthLogonCertTokenRequests          float64 `perflib:"OAuth Logon Certificate Token Requests"`
	OAuthPasswordGrantRequestFailures    float64 `perflib:"OAuth Password Grant Request Failures"`
	OAuthPasswordGrantRequests           float64 `perflib:"OAuth Password Grant Requests"`
	OAuthTokenRequests                   float64 `perflib:"OAuth Token Requests"`
	SAMLPTokenRequests                   float64 `perflib:"SAML-P Token Requests"`
	SSOAuthenticationFailures            float64 `perflib:"SSO Authentication Failures"`
	SSOAuthentications                   float64 `perflib:"SSO Authentications"`
	WSFedTokenRequests                   float64 `perflib:"WS-Fed Token Requests"`
	WSTrustTokenRequests                 float64 `perflib:"WS-Trust Token Requests"`
	UsernamePasswordAuthnFailures        float64 `perflib:"U/P Authentication Failures"`
	UsernamePasswordAuthentications      float64 `perflib:"U/P Authentications"`
	ExternalAuthentications              float64 `perflib:"External Authentications"`
	ExternalAuthNFailures                float64 `perflib:"External Authentication Failures"`
	ArtifactDBFailures                   float64 `perflib:"Artifact Database Connection Failures"`
	AvgArtifactDBQueryTime               float64 `perflib:"Average Artifact Database Query Time"`
	ConfigDBFailures                     float64 `perflib:"Configuration Database Connection Failures"`
	AvgConfigDBQueryTime                 float64 `perflib:"Average Config Database Query Time"`
	FederationMetadataRequests           float64 `perflib:"Federation Metadata Requests"`
}

func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var adfsData []perflibADFS

	err := perflib.UnmarshalObject(ctx.PerfObjects["AD FS"], &adfsData, logger)
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

	ch <- prometheus.MustNewConstMetric(
		c.oAuthAuthZRequests,
		prometheus.CounterValue,
		adfsData[0].OAuthAuthZRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientAuthentications,
		prometheus.CounterValue,
		adfsData[0].OAuthClientAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientAuthenticationsFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthClientAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientCredentialsRequestFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthClientCredentialRequestFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientCredentialsRequests,
		prometheus.CounterValue,
		adfsData[0].OAuthClientCredentialRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientPrivateKeyJwtAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthClientPrivKeyJWTAuthnFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientPrivateKeyJwtAuthentications,
		prometheus.CounterValue,
		adfsData[0].OAuthClientPrivKeyJWTAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretBasicAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthClientBasicAuthnFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretBasicAuthentications,
		prometheus.CounterValue,
		adfsData[0].OAuthClientBasicAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretPostAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthClientSecretPostAuthnFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientSecretPostAuthentications,
		prometheus.CounterValue,
		adfsData[0].OAuthClientSecretPostAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientWindowsIntegratedAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthClientWindowsAuthnFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthClientWindowsIntegratedAuthentications,
		prometheus.CounterValue,
		adfsData[0].OAuthClientWindowsAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthLogonCertificateRequestFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthLogonCertRequestFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthLogonCertificateTokenRequests,
		prometheus.CounterValue,
		adfsData[0].OAuthLogonCertTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthPasswordGrantRequestFailures,
		prometheus.CounterValue,
		adfsData[0].OAuthPasswordGrantRequestFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthPasswordGrantRequests,
		prometheus.CounterValue,
		adfsData[0].OAuthPasswordGrantRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.oAuthTokenRequests,
		prometheus.CounterValue,
		adfsData[0].OAuthTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.samlPTokenRequests,
		prometheus.CounterValue,
		adfsData[0].SAMLPTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ssoAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].SSOAuthenticationFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ssoAuthentications,
		prometheus.CounterValue,
		adfsData[0].SSOAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wsfedTokenRequests,
		prometheus.CounterValue,
		adfsData[0].WSFedTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wstrustTokenRequests,
		prometheus.CounterValue,
		adfsData[0].WSTrustTokenRequests,
	)

	ch <- prometheus.MustNewConstMetric(
		c.upAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].UsernamePasswordAuthnFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.upAuthentications,
		prometheus.CounterValue,
		adfsData[0].UsernamePasswordAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.externalAuthenticationFailures,
		prometheus.CounterValue,
		adfsData[0].ExternalAuthNFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.externalAuthentications,
		prometheus.CounterValue,
		adfsData[0].ExternalAuthentications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.artifactDBFailures,
		prometheus.CounterValue,
		adfsData[0].ArtifactDBFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.avgArtifactDBQueryTime,
		prometheus.CounterValue,
		adfsData[0].AvgArtifactDBQueryTime*math.Pow(10, -8),
	)

	ch <- prometheus.MustNewConstMetric(
		c.configDBFailures,
		prometheus.CounterValue,
		adfsData[0].ConfigDBFailures,
	)

	ch <- prometheus.MustNewConstMetric(
		c.avgConfigDBQueryTime,
		prometheus.CounterValue,
		adfsData[0].AvgConfigDBQueryTime*math.Pow(10, -8),
	)

	ch <- prometheus.MustNewConstMetric(
		c.federationMetadataRequests,
		prometheus.CounterValue,
		adfsData[0].FederationMetadataRequests,
	)

	return nil
}
