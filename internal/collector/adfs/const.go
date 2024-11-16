//go:build windows

package adfs

const (
	adLoginConnectionFailures                      = "AD Login Connection Failures"
	artifactDBFailures                             = "Artifact Database Connection Failures"
	avgArtifactDBQueryTime                         = "Average Artifact Database Query Time"
	avgConfigDBQueryTime                           = "Average Config Database Query Time"
	certificateAuthentications                     = "Certificate Authentications"
	configDBFailures                               = "Configuration Database Connection Failures"
	deviceAuthentications                          = "Device Authentications"
	externalAuthentications                        = "External Authentications"
	externalAuthNFailures                          = "External Authentication Failures"
	extranetAccountLockouts                        = "Extranet Account Lockouts"
	federatedAuthentications                       = "Federated Authentications"
	federationMetadataRequests                     = "Federation Metadata Requests"
	oAuthAuthZRequests                             = "OAuth AuthZ Requests"
	oAuthClientAuthenticationFailures              = "OAuth Client Authentications Failures"
	oAuthClientAuthentications                     = "OAuth Client Authentications"
	oAuthClientBasicAuthenticationFailures         = "OAuth Client Secret Basic Authentication Failures"
	oAuthClientBasicAuthentications                = "OAuth Client Secret Basic Authentication Requests"
	oAuthClientCredentialRequestFailures           = "OAuth Client Credentials Request Failures"
	oAuthClientCredentialRequests                  = "OAuth Client Credentials Requests"
	oAuthClientPrivateKeyJWTAuthenticationFailures = "OAuth Client Private Key Jwt Authentication Failures"
	oAuthClientPrivateKeyJWTAuthentications        = "OAuth Client Private Key Jwt Authentications"
	oAuthClientSecretPostAuthenticationFailures    = "OAuth Client Secret Post Authentication Failures"
	oAuthClientSecretPostAuthentications           = "OAuth Client Secret Post Authentications"
	oAuthClientWindowsAuthenticationFailures       = "OAuth Client Windows Integrated Authentication Failures"
	oAuthClientWindowsAuthentications              = "OAuth Client Windows Integrated Authentications"
	oAuthLogonCertRequestFailures                  = "OAuth Logon Certificate Request Failures"
	oAuthLogonCertTokenRequests                    = "OAuth Logon Certificate Token Requests"
	oAuthPasswordGrantRequestFailures              = "OAuth Password Grant Request Failures"
	oAuthPasswordGrantRequests                     = "OAuth Password Grant Requests"
	oAuthTokenRequests                             = "OAuth Token Requests"
	passiveRequests                                = "Passive Requests"
	passportAuthentications                        = "Microsoft Passport Authentications"
	passwordChangeFailed                           = "Password Change Failed Requests"
	passwordChangeSucceeded                        = "Password Change Successful Requests"
	samlPTokenRequests                             = "SAML-P Token Requests"
	ssoAuthenticationFailures                      = "SSO Authentication Failures"
	ssoAuthentications                             = "SSO Authentications"
	tokenRequests                                  = "Token Requests"
	usernamePasswordAuthenticationFailures         = "U/P Authentication Failures"
	usernamePasswordAuthentications                = "U/P Authentications"
	windowsIntegratedAuthentications               = "Windows Integrated Authentications"
	wsFedTokenRequests                             = "WS-Fed Token Requests"
	wsTrustTokenRequests                           = "WS-Trust Token Requests"
)

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
