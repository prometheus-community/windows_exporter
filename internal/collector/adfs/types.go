// Copyright 2024 The Prometheus Authors
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

type perfDataCounterValues struct {
	AdLoginConnectionFailures                      float64 `perfdata:"AD Login Connection Failures"`
	ArtifactDBFailures                             float64 `perfdata:"Artifact Database Connection Failures"`
	AvgArtifactDBQueryTime                         float64 `perfdata:"Average Artifact Database Query Time"`
	AvgConfigDBQueryTime                           float64 `perfdata:"Average Config Database Query Time"`
	CertificateAuthentications                     float64 `perfdata:"Certificate Authentications"`
	ConfigDBFailures                               float64 `perfdata:"Configuration Database Connection Failures"`
	DeviceAuthentications                          float64 `perfdata:"Device Authentications"`
	ExternalAuthentications                        float64 `perfdata:"External Authentications"`
	ExternalAuthNFailures                          float64 `perfdata:"External Authentication Failures"`
	ExtranetAccountLockouts                        float64 `perfdata:"Extranet Account Lockouts"`
	FederatedAuthentications                       float64 `perfdata:"Federated Authentications"`
	FederationMetadataRequests                     float64 `perfdata:"Federation Metadata Requests"`
	OAuthAuthZRequests                             float64 `perfdata:"OAuth AuthZ Requests"`
	OAuthClientAuthenticationFailures              float64 `perfdata:"OAuth Client Authentications Failures"`
	OAuthClientAuthentications                     float64 `perfdata:"OAuth Client Authentications"`
	OAuthClientBasicAuthenticationFailures         float64 `perfdata:"OAuth Client Secret Basic Authentication Failures"`
	OAuthClientBasicAuthentications                float64 `perfdata:"OAuth Client Secret Basic Authentications"`
	OAuthClientCredentialRequestFailures           float64 `perfdata:"OAuth Client Credentials Request Failures"`
	OAuthClientCredentialRequests                  float64 `perfdata:"OAuth Client Credentials Requests"`
	OAuthClientPrivateKeyJWTAuthenticationFailures float64 `perfdata:"OAuth Client Private Key Jwt Authentication Failures"`
	OAuthClientPrivateKeyJWTAuthentications        float64 `perfdata:"OAuth Client Private Key Jwt Authentications"`
	OAuthClientSecretPostAuthenticationFailures    float64 `perfdata:"OAuth Client Secret Post Authentication Failures"`
	OAuthClientSecretPostAuthentications           float64 `perfdata:"OAuth Client Secret Post Authentications"`
	OAuthClientWindowsAuthenticationFailures       float64 `perfdata:"OAuth Client Windows Integrated Authentication Failures"`
	OAuthClientWindowsAuthentications              float64 `perfdata:"OAuth Client Windows Integrated Authentications"`
	OAuthLogonCertRequestFailures                  float64 `perfdata:"OAuth Logon Certificate Request Failures"`
	OAuthLogonCertTokenRequests                    float64 `perfdata:"OAuth Logon Certificate Token Requests"`
	OAuthPasswordGrantRequestFailures              float64 `perfdata:"OAuth Password Grant Request Failures"`
	OAuthPasswordGrantRequests                     float64 `perfdata:"OAuth Password Grant Requests"`
	OAuthTokenRequests                             float64 `perfdata:"OAuth Token Requests"`
	PassiveRequests                                float64 `perfdata:"Passive Requests"`
	PassportAuthentications                        float64 `perfdata:"Microsoft Passport Authentications"`
	PasswordChangeFailed                           float64 `perfdata:"Password Change Failed Requests"`
	PasswordChangeSucceeded                        float64 `perfdata:"Password Change Successful Requests"`
	SamlPTokenRequests                             float64 `perfdata:"SAML-P Token Requests"`
	SsoAuthenticationFailures                      float64 `perfdata:"SSO Authentication Failures"`
	SsoAuthentications                             float64 `perfdata:"SSO Authentications"`
	TokenRequests                                  float64 `perfdata:"Token Requests"`
	UsernamePasswordAuthenticationFailures         float64 `perfdata:"U/P Authentication Failures"`
	UsernamePasswordAuthentications                float64 `perfdata:"U/P Authentications"`
	WindowsIntegratedAuthentications               float64 `perfdata:"Windows Integrated Authentications"`
	WsFedTokenRequests                             float64 `perfdata:"WS-Fed Token Requests"`
	WsTrustTokenRequests                           float64 `perfdata:"WS-Trust Token Requests"`
}
