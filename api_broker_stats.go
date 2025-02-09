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

// BrokerStats is admin interface for broker stats management
type BrokerStats interface {
	// GetMetrics returns Monitoring metrics
	GetMetrics() ([]Metrics, error)

	// GetMBeans requests JSON string server mbean dump
	GetMBeans() ([]Metrics, error)

	// GetTopics returns JSON string topics stats
	GetTopics() (string, error)

	// GetLoadReport returns load report of broker
	GetLoadReport() (*LocalBrokerData, error)

	// GetAllocatorStats returns stats from broker
	GetAllocatorStats(allocatorName string) (*AllocatorStats, error)
}

type brokerStats struct {
	pulsar     *pulsarClient
	basePath   string
	apiVersion APIVersion
}

// BrokerStats is used to access the broker stats endpoints
func (c *pulsarClient) BrokerStats() BrokerStats {
	return &brokerStats{
		pulsar:     c,
		basePath:   "/broker-stats",
		apiVersion: c.apiProfile.BrokerStats,
	}
}

func (bs *brokerStats) GetMetrics() ([]Metrics, error) {
	endpoint := bs.pulsar.endpoint(bs.apiVersion, bs.basePath, "/metrics")
	var response []Metrics
	err := bs.pulsar.restClient.Get(endpoint, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (bs *brokerStats) GetMBeans() ([]Metrics, error) {
	endpoint := bs.pulsar.endpoint(bs.apiVersion, bs.basePath, "/mbeans")
	var response []Metrics
	err := bs.pulsar.restClient.Get(endpoint, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (bs *brokerStats) GetTopics() (string, error) {
	endpoint := bs.pulsar.endpoint(bs.apiVersion, bs.basePath, "/topics")
	buf, err := bs.pulsar.restClient.GetWithQueryParams(endpoint, nil, nil, false)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (bs *brokerStats) GetLoadReport() (*LocalBrokerData, error) {
	endpoint := bs.pulsar.endpoint(bs.apiVersion, bs.basePath, "/load-report")
	response := NewLocalBrokerData()
	err := bs.pulsar.restClient.Get(endpoint, &response)
	if err != nil {
		return nil, nil
	}
	return &response, nil
}

func (bs *brokerStats) GetAllocatorStats(allocatorName string) (*AllocatorStats, error) {
	endpoint := bs.pulsar.endpoint(bs.apiVersion, bs.basePath, "/allocator-stats", allocatorName)
	var allocatorStats AllocatorStats
	err := bs.pulsar.restClient.Get(endpoint, &allocatorStats)
	if err != nil {
		return nil, err
	}
	return &allocatorStats, nil
}
