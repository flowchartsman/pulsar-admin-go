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

type ResourceQuotas interface {
	// Get default resource quota for new resource bundles.
	GetDefaultResourceQuota() (*ResourceQuota, error)

	// Set default resource quota for new namespace bundles.
	SetDefaultResourceQuota(quota ResourceQuota) error

	// Get resource quota of a namespace bundle.
	GetNamespaceBundleResourceQuota(namespace, bundle string) (*ResourceQuota, error)

	// Set resource quota for a namespace bundle.
	SetNamespaceBundleResourceQuota(namespace, bundle string, quota ResourceQuota) error

	// Reset resource quota for a namespace bundle to default value.
	ResetNamespaceBundleResourceQuota(namespace, bundle string) error
}

type resource struct {
	pulsar     *pulsarClient
	basePath   string
	apiVersion APIVersion
}

func (c *pulsarClient) ResourceQuotas() ResourceQuotas {
	return &resource{
		pulsar:     c,
		basePath:   "/resource-quotas",
		apiVersion: c.apiProfile.ResourceQuotas,
	}
}

func (r *resource) GetDefaultResourceQuota() (*ResourceQuota, error) {
	endpoint := r.pulsar.endpoint(r.apiVersion, r.basePath)
	var quota ResourceQuota
	err := r.pulsar.restClient.Get(endpoint, &quota)
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *resource) SetDefaultResourceQuota(quota ResourceQuota) error {
	endpoint := r.pulsar.endpoint(r.apiVersion, r.basePath)
	return r.pulsar.restClient.Post(endpoint, &quota)
}

func (r *resource) GetNamespaceBundleResourceQuota(namespace, bundle string) (*ResourceQuota, error) {
	endpoint := r.pulsar.endpoint(r.apiVersion, r.basePath, namespace, bundle)
	var quota ResourceQuota
	err := r.pulsar.restClient.Get(endpoint, &quota)
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *resource) SetNamespaceBundleResourceQuota(namespace, bundle string, quota ResourceQuota) error {
	endpoint := r.pulsar.endpoint(r.apiVersion, r.basePath, namespace, bundle)
	return r.pulsar.restClient.Post(endpoint, &quota)
}

func (r *resource) ResetNamespaceBundleResourceQuota(namespace, bundle string) error {
	endpoint := r.pulsar.endpoint(r.apiVersion, r.basePath, namespace, bundle)
	return r.pulsar.restClient.Delete(endpoint)
}
