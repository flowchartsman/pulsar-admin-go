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

// Tenants is admin interface for tenants management
type Tenants interface {
	// Create a new tenant
	Create(TenantData) error

	// Delete an existing tenant
	Delete(string) error

	// Update the admins for a tenant
	Update(TenantData) error

	// List returns the list of tenants
	List() ([]string, error)

	// Get returns the config of the tenant.
	Get(string) (TenantData, error)
}

type tenants struct {
	pulsar     *pulsarClient
	basePath   string
	apiVersion APIVersion
}

// Tenants is used to access the tenants endpoints
func (c *pulsarClient) Tenants() Tenants {
	return &tenants{
		pulsar:     c,
		basePath:   "/tenants",
		apiVersion: c.apiProfile.Tenants,
	}
}

func (c *tenants) Create(data TenantData) error {
	endpoint := c.pulsar.endpoint(c.apiVersion, c.basePath, data.Name)
	return c.pulsar.restClient.Put(endpoint, &data)
}

func (c *tenants) Delete(name string) error {
	endpoint := c.pulsar.endpoint(c.apiVersion, c.basePath, name)
	return c.pulsar.restClient.Delete(endpoint)
}

func (c *tenants) Update(data TenantData) error {
	endpoint := c.pulsar.endpoint(c.apiVersion, c.basePath, data.Name)
	return c.pulsar.restClient.Post(endpoint, &data)
}

func (c *tenants) List() ([]string, error) {
	var tenantList []string
	endpoint := c.pulsar.endpoint(c.apiVersion, c.basePath, "")
	err := c.pulsar.restClient.Get(endpoint, &tenantList)
	return tenantList, err
}

func (c *tenants) Get(name string) (TenantData, error) {
	var data TenantData
	endpoint := c.pulsar.endpoint(c.apiVersion, c.basePath, name)
	err := c.pulsar.restClient.Get(endpoint, &data)
	return data, err
}
