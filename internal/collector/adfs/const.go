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
	oAuthClientBasicAuthentications                = "OAuth Client Secret Basic Authentications"
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
