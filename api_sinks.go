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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// Sinks is admin interface for sinks management
type Sinks interface {
	// ListSinks returns the list of all the Pulsar Sinks.
	ListSinks(tenant, namespace string) ([]string, error)

	// GetSink returns the configuration for the specified sink
	GetSink(tenant, namespace, Sink string) (SinkConfig, error)

	// CreateSink creates a new sink
	CreateSink(config *SinkConfig, fileName string) error

	// CreateSinkWithURL creates a new sink by providing url from which fun-pkg can be downloaded. supported url: http/file
	CreateSinkWithURL(config *SinkConfig, pkgURL string) error

	// UpdateSink updates the configuration for a sink.
	UpdateSink(config *SinkConfig, fileName string, options *UpdateOptions) error

	// UpdateSinkWithURL updates a sink by providing url from which fun-pkg can be downloaded. supported url: http/file
	UpdateSinkWithURL(config *SinkConfig, pkgURL string, options *UpdateOptions) error

	// DeleteSink deletes an existing sink
	DeleteSink(tenant, namespace, Sink string) error

	// GetSinkStatus returns the current status of a sink.
	GetSinkStatus(tenant, namespace, Sink string) (SinkStatus, error)

	// GetSinkStatusWithID returns the current status of a sink instance.
	GetSinkStatusWithID(tenant, namespace, Sink string, id int) (SinkInstanceStatusData, error)

	// RestartSink restarts all sink instances
	RestartSink(tenant, namespace, Sink string) error

	// RestartSinkWithID restarts sink instance
	RestartSinkWithID(tenant, namespace, Sink string, id int) error

	// StopSink stops all sink instances
	StopSink(tenant, namespace, Sink string) error

	// StopSinkWithID stops sink instance
	StopSinkWithID(tenant, namespace, Sink string, id int) error

	// StartSink starts all sink instances
	StartSink(tenant, namespace, Sink string) error

	// StartSinkWithID starts sink instance
	StartSinkWithID(tenant, namespace, Sink string, id int) error

	// GetBuiltInSinks fetches a list of supported Pulsar IO sinks currently running in cluster mode
	GetBuiltInSinks() ([]*ConnectorDefinition, error)

	// ReloadBuiltInSinks reload the available built-in connectors, include Source and Sink
	ReloadBuiltInSinks() error
}

type sinks struct {
	pulsar     *pulsarClient
	basePath   string
	apiVersion APIVersion
}

// Sinks is used to access the sinks endpoints
func (c *pulsarClient) Sinks() Sinks {
	return &sinks{
		pulsar:     c,
		basePath:   "/sinks",
		apiVersion: c.apiProfile.Sinks,
	}
}

func (s *sinks) createStringFromField(w *multipart.Writer, value string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s" `, value))
	h.Set("Content-Type", "application/json")
	return w.CreatePart(h)
}

func (s *sinks) createTextFromFiled(w *multipart.Writer, value string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s" `, value))
	h.Set("Content-Type", "text/plain")
	return w.CreatePart(h)
}

func (s *sinks) ListSinks(tenant, namespace string) ([]string, error) {
	var sinks []string
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace)
	err := s.pulsar.restClient.Get(endpoint, &sinks)
	return sinks, err
}

func (s *sinks) GetSink(tenant, namespace, sink string) (SinkConfig, error) {
	var sinkConfig SinkConfig
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink)
	err := s.pulsar.restClient.Get(endpoint, &sinkConfig)
	return sinkConfig, err
}

func (s *sinks) CreateSink(config *SinkConfig, fileName string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, config.Tenant, config.Namespace, config.Name)

	// buffer to store our request as bytes
	bodyBuf := bytes.NewBufferString("")

	multiPartWriter := multipart.NewWriter(bodyBuf)
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	stringWriter, err := s.createStringFromField(multiPartWriter, "sinkConfig")
	if err != nil {
		return err
	}

	_, err = stringWriter.Write(jsonData)
	if err != nil {
		return err
	}

	if fileName != "" && !strings.HasPrefix(fileName, "builtin://") {
		// If the function code is built in, we don't need to submit here
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		part, err := multiPartWriter.CreateFormFile("data", filepath.Base(file.Name()))
		if err != nil {
			return err
		}

		// copy the actual file content to the filed's writer
		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}

	// In here, we completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	if err = multiPartWriter.Close(); err != nil {
		return err
	}

	contentType := multiPartWriter.FormDataContentType()
	err = s.pulsar.restClient.PostWithMultiPart(endpoint, nil, bodyBuf, contentType)
	if err != nil {
		return err
	}

	return nil
}

func (s *sinks) CreateSinkWithURL(config *SinkConfig, pkgURL string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, config.Tenant, config.Namespace, config.Name)
	// buffer to store our request as bytes
	bodyBuf := bytes.NewBufferString("")

	multiPartWriter := multipart.NewWriter(bodyBuf)

	textWriter, err := s.createTextFromFiled(multiPartWriter, "url")
	if err != nil {
		return err
	}

	_, err = textWriter.Write([]byte(pkgURL))
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	stringWriter, err := s.createStringFromField(multiPartWriter, "sinkConfig")
	if err != nil {
		return err
	}

	_, err = stringWriter.Write(jsonData)
	if err != nil {
		return err
	}

	if err = multiPartWriter.Close(); err != nil {
		return err
	}

	contentType := multiPartWriter.FormDataContentType()
	err = s.pulsar.restClient.PostWithMultiPart(endpoint, nil, bodyBuf, contentType)
	if err != nil {
		return err
	}

	return nil
}

func (s *sinks) UpdateSink(config *SinkConfig, fileName string, updateOptions *UpdateOptions) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, config.Tenant, config.Namespace, config.Name)
	// buffer to store our request as bytes
	bodyBuf := bytes.NewBufferString("")

	multiPartWriter := multipart.NewWriter(bodyBuf)

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	stringWriter, err := s.createStringFromField(multiPartWriter, "sinkConfig")
	if err != nil {
		return err
	}

	_, err = stringWriter.Write(jsonData)
	if err != nil {
		return err
	}

	if updateOptions != nil {
		updateData, err := json.Marshal(updateOptions)
		if err != nil {
			return err
		}

		updateStrWriter, err := s.createStringFromField(multiPartWriter, "updateOptions")
		if err != nil {
			return err
		}

		_, err = updateStrWriter.Write(updateData)
		if err != nil {
			return err
		}
	}

	if fileName != "" && !strings.HasPrefix(fileName, "builtin://") {
		// If the function code is built in, we don't need to submit here
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		part, err := multiPartWriter.CreateFormFile("data", filepath.Base(file.Name()))
		if err != nil {
			return err
		}

		// copy the actual file content to the filed's writer
		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}

	// In here, we completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	if err = multiPartWriter.Close(); err != nil {
		return err
	}

	contentType := multiPartWriter.FormDataContentType()
	err = s.pulsar.restClient.PutWithMultiPart(endpoint, bodyBuf, contentType)
	if err != nil {
		return err
	}

	return nil
}

func (s *sinks) UpdateSinkWithURL(config *SinkConfig, pkgURL string, updateOptions *UpdateOptions) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, config.Tenant, config.Namespace, config.Name)
	// buffer to store our request as bytes
	bodyBuf := bytes.NewBufferString("")

	multiPartWriter := multipart.NewWriter(bodyBuf)

	textWriter, err := s.createTextFromFiled(multiPartWriter, "url")
	if err != nil {
		return err
	}

	_, err = textWriter.Write([]byte(pkgURL))
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	stringWriter, err := s.createStringFromField(multiPartWriter, "sinkConfig")
	if err != nil {
		return err
	}

	_, err = stringWriter.Write(jsonData)
	if err != nil {
		return err
	}

	if updateOptions != nil {
		updateData, err := json.Marshal(updateOptions)
		if err != nil {
			return err
		}

		updateStrWriter, err := s.createStringFromField(multiPartWriter, "updateOptions")
		if err != nil {
			return err
		}

		_, err = updateStrWriter.Write(updateData)
		if err != nil {
			return err
		}
	}

	// In here, we completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	if err = multiPartWriter.Close(); err != nil {
		return err
	}

	contentType := multiPartWriter.FormDataContentType()
	err = s.pulsar.restClient.PutWithMultiPart(endpoint, bodyBuf, contentType)
	if err != nil {
		return err
	}

	return nil
}

func (s *sinks) DeleteSink(tenant, namespace, sink string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink)
	return s.pulsar.restClient.Delete(endpoint)
}

func (s *sinks) GetSinkStatus(tenant, namespace, sink string) (SinkStatus, error) {
	var sinkStatus SinkStatus
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink)
	err := s.pulsar.restClient.Get(endpoint+"/status", &sinkStatus)
	return sinkStatus, err
}

func (s *sinks) GetSinkStatusWithID(tenant, namespace, sink string, id int) (SinkInstanceStatusData, error) {
	var sinkInstanceStatusData SinkInstanceStatusData
	instanceID := fmt.Sprintf("%d", id)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink, instanceID)
	err := s.pulsar.restClient.Get(endpoint+"/status", &sinkInstanceStatusData)
	return sinkInstanceStatusData, err
}

func (s *sinks) RestartSink(tenant, namespace, sink string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink)
	return s.pulsar.restClient.Post(endpoint+"/restart", nil)
}

func (s *sinks) RestartSinkWithID(tenant, namespace, sink string, instanceID int) error {
	id := fmt.Sprintf("%d", instanceID)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink, id)

	return s.pulsar.restClient.Post(endpoint+"/restart", nil)
}

func (s *sinks) StopSink(tenant, namespace, sink string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink)
	return s.pulsar.restClient.Post(endpoint+"/stop", nil)
}

func (s *sinks) StopSinkWithID(tenant, namespace, sink string, instanceID int) error {
	id := fmt.Sprintf("%d", instanceID)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink, id)

	return s.pulsar.restClient.Post(endpoint+"/stop", nil)
}

func (s *sinks) StartSink(tenant, namespace, sink string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink)
	return s.pulsar.restClient.Post(endpoint+"/start", nil)
}

func (s *sinks) StartSinkWithID(tenant, namespace, sink string, instanceID int) error {
	id := fmt.Sprintf("%d", instanceID)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, sink, id)

	return s.pulsar.restClient.Post(endpoint+"/start", nil)
}

func (s *sinks) GetBuiltInSinks() ([]*ConnectorDefinition, error) {
	var connectorDefinition []*ConnectorDefinition
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, "builtinsinks")
	err := s.pulsar.restClient.Get(endpoint, &connectorDefinition)
	return connectorDefinition, err
}

func (s *sinks) ReloadBuiltInSinks() error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, "reloadBuiltInSinks")
	return s.pulsar.restClient.Post(endpoint, nil)
}
