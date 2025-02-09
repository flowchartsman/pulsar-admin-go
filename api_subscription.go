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
	"encoding/binary"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Subscriptions is admin interface for subscriptions management
type Subscriptions interface {
	// Create a new subscription on a topic
	Create(TopicName, string, MessageID) error

	// Delete a subscription.
	// Delete a persistent subscription from a topic. There should not be any active consumers on the subscription
	Delete(TopicName, string) error

	// ForceDelete deletes a subscription forcefully
	ForceDelete(TopicName, string) error

	// List returns the list of subscriptions
	List(TopicName) ([]string, error)

	// ResetCursorToMessageID resets cursor position on a topic subscription
	// @param
	// messageID reset subscription to messageId (or previous nearest messageId if given messageId is not valid)
	ResetCursorToMessageID(TopicName, string, MessageID) error

	// ResetCursorToTimestamp resets cursor position on a topic subscription
	// @param
	// time reset subscription to position closest to time in ms since epoch
	ResetCursorToTimestamp(TopicName, string, int64) error

	// ClearBacklog skips all messages on a topic subscription
	ClearBacklog(TopicName, string) error

	// SkipMessages skips messages on a topic subscription
	SkipMessages(TopicName, string, int64) error

	// ExpireMessages expires all messages older than given N (expireTimeInSeconds) seconds for a given subscription
	ExpireMessages(TopicName, string, int64) error

	// ExpireAllMessages expires all messages older than given N (expireTimeInSeconds) seconds for all
	// subscriptions of the persistent-topic
	ExpireAllMessages(TopicName, int64) error

	// PeekMessages peeks messages from a topic subscription
	PeekMessages(TopicName, string, int) ([]*Message, error)

	// GetMessageByID gets message by its ledgerID and entryID
	GetMessageByID(topic TopicName, ledgerID, entryID int64) (*Message, error)
}

type subscriptions struct {
	pulsar   *pulsarClient
	basePath string
	SubPath  string
	topicAPI APIVersion
}

// Subscriptions is used to access the subscriptions endpoints
func (c *pulsarClient) Subscriptions() Subscriptions {
	return &subscriptions{
		pulsar:   c,
		basePath: "",
		SubPath:  "subscription",
		topicAPI: c.apiProfile.Topics,
	}
}

func (s *subscriptions) Create(topic TopicName, sName string, messageID MessageID) error {
	endpoint := s.pulsar.endpoint(s.topicAPI, s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(sName))
	return s.pulsar.restClient.Put(endpoint, messageID)
}

func (s *subscriptions) delete(topic TopicName, subName string, force bool) error {
	endpoint := s.pulsar.endpoint(s.topicAPI, s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(subName))
	queryParams := make(map[string]string)
	queryParams["force"] = strconv.FormatBool(force)
	return s.pulsar.restClient.DeleteWithQueryParams(endpoint, queryParams)
}

func (s *subscriptions) Delete(topic TopicName, sName string) error {
	return s.delete(topic, sName, false)
}

func (s *subscriptions) ForceDelete(topic TopicName, sName string) error {
	return s.delete(topic, sName, true)
}

func (s *subscriptions) List(topic TopicName) ([]string, error) {
	endpoint := s.pulsar.endpoint(s.topicAPI, s.basePath, topic.GetRestPath(), "subscriptions")
	var list []string
	return list, s.pulsar.restClient.Get(endpoint, &list)
}

func (s *subscriptions) ResetCursorToMessageID(topic TopicName, sName string, id MessageID) error {
	endpoint := s.pulsar.endpoint(s.topicAPI, s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(sName),
		"resetcursor")
	return s.pulsar.restClient.Post(endpoint, id)
}

func (s *subscriptions) ResetCursorToTimestamp(topic TopicName, sName string, timestamp int64) error {
	endpoint := s.pulsar.endpoint(s.topicAPI,
		s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(sName),
		"resetcursor", strconv.FormatInt(timestamp, 10))
	return s.pulsar.restClient.Post(endpoint, nil)
}

func (s *subscriptions) ClearBacklog(topic TopicName, sName string) error {
	endpoint := s.pulsar.endpoint(s.topicAPI,
		s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(sName), "skip_all")
	return s.pulsar.restClient.Post(endpoint, nil)
}

func (s *subscriptions) SkipMessages(topic TopicName, sName string, n int64) error {
	endpoint := s.pulsar.endpoint(s.topicAPI,
		s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(sName),
		"skip", strconv.FormatInt(n, 10))
	return s.pulsar.restClient.Post(endpoint, nil)
}

func (s *subscriptions) ExpireMessages(topic TopicName, sName string, expire int64) error {
	endpoint := s.pulsar.endpoint(s.topicAPI,
		s.basePath, topic.GetRestPath(), s.SubPath, url.PathEscape(sName),
		"expireMessages", strconv.FormatInt(expire, 10))
	return s.pulsar.restClient.Post(endpoint, nil)
}

func (s *subscriptions) ExpireAllMessages(topic TopicName, expire int64) error {
	endpoint := s.pulsar.endpoint(s.topicAPI,
		s.basePath, topic.GetRestPath(), "all_subscription",
		"expireMessages", strconv.FormatInt(expire, 10))
	return s.pulsar.restClient.Post(endpoint, nil)
}

func (s *subscriptions) PeekMessages(topic TopicName, sName string, n int) ([]*Message, error) {
	var msgs []*Message

	count := 1
	for n > 0 {
		m, err := s.peekNthMessage(topic, sName, count)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m...)
		n -= len(m)
		count++
	}

	return msgs, nil
}

func (s *subscriptions) peekNthMessage(topic TopicName, sName string, pos int) ([]*Message, error) {
	endpoint := s.pulsar.endpoint(s.topicAPI, s.basePath, topic.GetRestPath(), "subscription", url.PathEscape(sName),
		"position", strconv.Itoa(pos))

	resp, err := s.pulsar.restClient.MakeRequest(http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	defer safeRespClose(resp)

	return handleResp(topic, resp)
}

func (s *subscriptions) GetMessageByID(topic TopicName, ledgerID, entryID int64) (*Message, error) {
	ledgerIDStr := strconv.FormatInt(ledgerID, 10)
	entryIDStr := strconv.FormatInt(entryID, 10)

	endpoint := s.pulsar.endpoint(s.topicAPI, s.basePath, topic.GetRestPath(), "ledger", ledgerIDStr, "entry", entryIDStr)
	resp, err := s.pulsar.restClient.MakeRequest(http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	defer safeRespClose(resp)

	messages, err := handleResp(topic, resp)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, nil
	}
	return messages[0], nil
}

// safeRespClose is used to close a response body
func safeRespClose(resp *http.Response) {
	if resp != nil {
		// ignore error since it is closing a response body
		_ = resp.Body.Close()
	}
}

const (
	PublishTimeHeader = "X-Pulsar-Publish-Time"
	BatchHeader       = "X-Pulsar-Num-Batch-Message"
	PropertyPrefix    = "X-Pulsar-Property-"
)

func handleResp(topic TopicName, resp *http.Response) ([]*Message, error) {
	msgID := resp.Header.Get("X-Pulsar-Message-ID")
	ID, err := ParseMessageID(msgID)
	if err != nil {
		return nil, err
	}

	// read data
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	properties := make(map[string]string)
	for k := range resp.Header {
		switch {
		case k == PublishTimeHeader:
			h := resp.Header.Get(k)
			if h != "" {
				properties["publish-time"] = h
			}
		case k == BatchHeader:
			h := resp.Header.Get(k)
			if h != "" {
				properties[BatchHeader] = h
			}
			return getIndividualMsgsFromBatch(topic, ID, payload, properties)
		case strings.Contains(k, PropertyPrefix):
			key := strings.TrimPrefix(k, PropertyPrefix)
			properties[key] = resp.Header.Get(k)
		}
	}

	return []*Message{NewMessage(topic.String(), *ID, payload, properties)}, nil
}

func getIndividualMsgsFromBatch(topic TopicName, msgID *MessageID, data []byte,
	properties map[string]string,
) ([]*Message, error) {
	batchSize, err := strconv.Atoi(properties[BatchHeader])
	if err != nil {
		return nil, nil
	}

	msgs := make([]*Message, 0, batchSize)

	// read all messages in batch
	buf32 := make([]byte, 4)
	rdBuf := bytes.NewReader(data)
	for i := 0; i < batchSize; i++ {
		msgID.BatchIndex = i
		// singleMetaSize
		if _, err := io.ReadFull(rdBuf, buf32); err != nil {
			return nil, err
		}
		singleMetaSize := binary.BigEndian.Uint32(buf32)

		// singleMeta
		singleMetaBuf := make([]byte, singleMetaSize)
		if _, err := io.ReadFull(rdBuf, singleMetaBuf); err != nil {
			return nil, err
		}

		singleMeta := new(SingleMessageMetadata)
		if err := proto.Unmarshal(singleMetaBuf, singleMeta); err != nil {
			return nil, err
		}

		if len(singleMeta.Properties) > 0 {
			for _, v := range singleMeta.Properties {
				k := *v.Key
				property := *v.Value
				properties[k] = property
			}
		}

		// payload
		singlePayload := make([]byte, singleMeta.GetPayloadSize())
		if _, err := io.ReadFull(rdBuf, singlePayload); err != nil {
			return nil, err
		}

		msgs = append(msgs, &Message{
			Topic:      topic.String(),
			MessageID:  *msgID,
			Payload:    singlePayload,
			Properties: properties,
		})
	}

	return msgs, nil
}
