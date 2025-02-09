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

// Sources is admin interface for sources management
type Sources interface {
	// ListSources returns the list of all the Pulsar Sources.
	ListSources(tenant, namespace string) ([]string, error)

	// GetSource return the configuration for the specified source
	GetSource(tenant, namespace, source string) (SourceConfig, error)

	// CreateSource creates a new source
	CreateSource(config *SourceConfig, fileName string) error

	// CreateSourceWithURL creates a new source by providing url from which fun-pkg can be downloaded.
	// supported url: http/file
	CreateSourceWithURL(config *SourceConfig, pkgURL string) error

	// UpdateSource updates the configuration for a source.
	UpdateSource(config *SourceConfig, fileName string, options *UpdateOptions) error

	// UpdateSourceWithURL updates a source by providing url from which fun-pkg can be downloaded. supported url: http/file
	UpdateSourceWithURL(config *SourceConfig, pkgURL string, options *UpdateOptions) error

	// DeleteSource deletes an existing source
	DeleteSource(tenant, namespace, source string) error

	// GetSourceStatus returns the current status of a source.
	GetSourceStatus(tenant, namespace, source string) (SourceStatus, error)

	// GetSourceStatusWithID returns the current status of a source instance.
	GetSourceStatusWithID(tenant, namespace, source string, id int) (SourceInstanceStatusData, error)

	// RestartSource restarts all source instances
	RestartSource(tenant, namespace, source string) error

	// RestartSourceWithID restarts source instance
	RestartSourceWithID(tenant, namespace, source string, id int) error

	// StopSource stops all source instances
	StopSource(tenant, namespace, source string) error

	// StopSourceWithID stops source instance
	StopSourceWithID(tenant, namespace, source string, id int) error

	// StartSource starts all source instances
	StartSource(tenant, namespace, source string) error

	// StartSourceWithID starts source instance
	StartSourceWithID(tenant, namespace, source string, id int) error

	// GetBuiltInSources fetches a list of supported Pulsar IO sources currently running in cluster mode
	GetBuiltInSources() ([]*ConnectorDefinition, error)

	// ReloadBuiltInSources reloads the available built-in connectors, include Source and Sink
	ReloadBuiltInSources() error
}

type sources struct {
	pulsar     *pulsarClient
	basePath   string
	apiVersion APIVersion
}

// Sources is used to access the sources endpoints
func (c *pulsarClient) Sources() Sources {
	return &sources{
		pulsar:     c,
		basePath:   "/sources",
		apiVersion: c.apiProfile.Sources,
	}
}

func (s *sources) createStringFromField(w *multipart.Writer, value string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s" `, value))
	h.Set("Content-Type", "application/json")
	return w.CreatePart(h)
}

func (s *sources) createTextFromFiled(w *multipart.Writer, value string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s" `, value))
	h.Set("Content-Type", "text/plain")
	return w.CreatePart(h)
}

func (s *sources) ListSources(tenant, namespace string) ([]string, error) {
	var sources []string
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace)
	err := s.pulsar.restClient.Get(endpoint, &sources)
	return sources, err
}

func (s *sources) GetSource(tenant, namespace, source string) (SourceConfig, error) {
	var sourceConfig SourceConfig
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source)
	err := s.pulsar.restClient.Get(endpoint, &sourceConfig)
	return sourceConfig, err
}

func (s *sources) CreateSource(config *SourceConfig, fileName string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, config.Tenant, config.Namespace, config.Name)

	// buffer to store our request as bytes
	bodyBuf := bytes.NewBufferString("")

	multiPartWriter := multipart.NewWriter(bodyBuf)
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	stringWriter, err := s.createStringFromField(multiPartWriter, "sourceConfig")
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

func (s *sources) CreateSourceWithURL(config *SourceConfig, pkgURL string) error {
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

	stringWriter, err := s.createStringFromField(multiPartWriter, "sourceConfig")
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

func (s *sources) UpdateSource(config *SourceConfig, fileName string, updateOptions *UpdateOptions) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, config.Tenant, config.Namespace, config.Name)
	// buffer to store our request as bytes
	bodyBuf := bytes.NewBufferString("")

	multiPartWriter := multipart.NewWriter(bodyBuf)

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	stringWriter, err := s.createStringFromField(multiPartWriter, "sourceConfig")
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

func (s *sources) UpdateSourceWithURL(config *SourceConfig, pkgURL string,
	updateOptions *UpdateOptions,
) error {
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

	stringWriter, err := s.createStringFromField(multiPartWriter, "sourceConfig")
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

func (s *sources) DeleteSource(tenant, namespace, source string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source)
	return s.pulsar.restClient.Delete(endpoint)
}

func (s *sources) GetSourceStatus(tenant, namespace, source string) (SourceStatus, error) {
	var sourceStatus SourceStatus
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source)
	err := s.pulsar.restClient.Get(endpoint+"/status", &sourceStatus)
	return sourceStatus, err
}

func (s *sources) GetSourceStatusWithID(tenant, namespace, source string, id int) (
	SourceInstanceStatusData, error,
) {
	var sourceInstanceStatusData SourceInstanceStatusData
	instanceID := fmt.Sprintf("%d", id)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source, instanceID)
	err := s.pulsar.restClient.Get(endpoint+"/status", &sourceInstanceStatusData)
	return sourceInstanceStatusData, err
}

func (s *sources) RestartSource(tenant, namespace, source string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source)
	return s.pulsar.restClient.Post(endpoint+"/restart", nil)
}

func (s *sources) RestartSourceWithID(tenant, namespace, source string, instanceID int) error {
	id := fmt.Sprintf("%d", instanceID)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source, id)

	return s.pulsar.restClient.Post(endpoint+"/restart", nil)
}

func (s *sources) StopSource(tenant, namespace, source string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source)
	return s.pulsar.restClient.Post(endpoint+"/stop", nil)
}

func (s *sources) StopSourceWithID(tenant, namespace, source string, instanceID int) error {
	id := fmt.Sprintf("%d", instanceID)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source, id)

	return s.pulsar.restClient.Post(endpoint+"/stop", nil)
}

func (s *sources) StartSource(tenant, namespace, source string) error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source)
	return s.pulsar.restClient.Post(endpoint+"/start", nil)
}

func (s *sources) StartSourceWithID(tenant, namespace, source string, instanceID int) error {
	id := fmt.Sprintf("%d", instanceID)
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, tenant, namespace, source, id)

	return s.pulsar.restClient.Post(endpoint+"/start", nil)
}

func (s *sources) GetBuiltInSources() ([]*ConnectorDefinition, error) {
	var connectorDefinition []*ConnectorDefinition
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, "builtinsources")
	err := s.pulsar.restClient.Get(endpoint, &connectorDefinition)
	return connectorDefinition, err
}

func (s *sources) ReloadBuiltInSources() error {
	endpoint := s.pulsar.endpoint(s.apiVersion, s.basePath, "reloadBuiltInSources")
	return s.pulsar.restClient.Post(endpoint, nil)
}
