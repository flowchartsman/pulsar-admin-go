// Copyright 2023 StreamNative, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package pulsaradmin

import "net/http"

type ClientConfig struct {
	// the web service url that pulsarctl connects to. Default is http://localhost:8080
	WebServiceURL string
	// the bookkeeper service url that pulsarctl connects to.
	BKWebServiceURL string
	// TLS Config
	TLSConfig TLSConfig
	// optional auth provider
	AuthProvider AuthProvider
	// optional custom HTTP transport
	CustomTransport *http.Transport
	// optional custom API profile to use different versions of different APIs
	APIProfile *APIProfile
}

type TLSConfig struct {
	TrustCertsFilePath         string
	AllowInsecureConnection    bool
	EnableHostnameVerification bool
}

type APIVersion int

const (
	undefined APIVersion = iota
	APIV1
	APIV2
	APIV3
)

const DefaultAPIVersion = "v2"

func (v APIVersion) String() string {
	switch v {
	case undefined:
		return DefaultAPIVersion
	case APIV1:
		return ""
	case APIV2:
		return "v2"
	case APIV3:
		return "v3"
	}

	return DefaultAPIVersion
}

type APIProfile struct {
	Clusters          APIVersion
	Functions         APIVersion
	Tenants           APIVersion
	Topics            APIVersion
	Sources           APIVersion
	Sinks             APIVersion
	Namespaces        APIVersion
	Schemas           APIVersion
	NsIsolationPolicy APIVersion
	Brokers           APIVersion
	BrokerStats       APIVersion
	ResourceQuotas    APIVersion
	FunctionsWorker   APIVersion
	Packages          APIVersion
}

func defaultAPIProfile() *APIProfile {
	return &APIProfile{
		Functions: APIV3,
	}
}
