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

type SourceStatus struct {
	NumInstances int                     `json:"numInstances"`
	NumRunning   int                     `json:"numRunning"`
	Instances    []*SourceInstanceStatus `json:"instances"`
}

type SourceInstanceStatus struct {
	InstanceID int                      `json:"instanceId"`
	Status     SourceInstanceStatusData `json:"status"`
}

type SourceInstanceStatusData struct {
	Running                bool                   `json:"running"`
	Err                    string                 `json:"error"`
	NumRestarts            int64                  `json:"numRestarts"`
	NumReceivedFromSource  int64                  `json:"numReceivedFromSource"`
	NumSystemExceptions    int64                  `json:"numSystemExceptions"`
	LatestSystemExceptions []ExceptionInformation `json:"latestSystemExceptions"`
	NumSourceExceptions    int64                  `json:"numSourceExceptions"`
	LatestSourceExceptions []ExceptionInformation `json:"latestSourceExceptions"`
	NumWritten             int64                  `json:"numWritten"`
	LastReceivedTime       int64                  `json:"lastReceivedTime"`
	WorkerID               string                 `json:"workerId"`
}
