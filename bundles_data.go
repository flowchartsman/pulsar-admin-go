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

type BundlesData struct {
	Boundaries []string `json:"boundaries"`
	NumBundles int      `json:"numBundles"`
}

func NewBundlesData(boundaries []string) BundlesData {
	return BundlesData{
		Boundaries: boundaries,
		NumBundles: len(boundaries) - 1,
	}
}

func NewBundlesDataWithNumBundles(numBundles int) *BundlesData {
	return &BundlesData{
		Boundaries: nil,
		NumBundles: numBundles,
	}
}

func NewDefaultBoundle() *BundlesData {
	bundleData := NewBundlesDataWithNumBundles(1)
	bundleData.Boundaries = append(bundleData.Boundaries, FirstBoundary, LastBoundary)
	return bundleData
}
