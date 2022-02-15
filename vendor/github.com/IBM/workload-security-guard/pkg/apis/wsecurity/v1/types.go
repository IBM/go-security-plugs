/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Guardian is a specification for a Guaerdian resource
type Guardian struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *GuardianSpec `json:"spec"`
	//	Status GuardianStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// GuardianList is a list of Guaerdian resources
type GuardianList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Guardian `json:"items"`
}

type GuardianSpec WsGate

/*
type U8MinmaxSlice []U8Minmax
type Uint64Slice []uint64
type void struct{}

type U8Minmax struct {
	Min uint8 `json:"min"`
	Max uint8 `json:"max"`
}
type SimpleValConfig struct { // 16 bytes
	Flags uint64 `json:"flags"`
	//Counters     [8]U8MinmaxSlice
	Runes        U8MinmaxSlice `json:"runes"`
	Digits       U8MinmaxSlice `json:"digits"`
	Letters      U8MinmaxSlice `json:"letters"`
	SpecialChars U8MinmaxSlice `json:"schars"`
	Words        U8MinmaxSlice `json:"words"`
	Numbers      U8MinmaxSlice `json:"numbers"`
	UnicodeFlags Uint64Slice   `json:"unicodeFlags"` //[]uint64
}
type KeyValConfig struct {
	Vals          map[string]*SimpleValConfig `json:"vals"`          // Profile the value of whitelisted keys
	MinimalSet    map[string]void             `json:"minimalSet"`    // Mandatory keys
	OtherVals     *SimpleValConfig            `json:"otherVals"`     // Profile the values of other keys
	OtherKeynames *SimpleValConfig            `json:"otherKeynames"` // Profile the keynames of other keys
}
type UrlConfig struct {
	Val      SimpleValConfig `json:"val"`
	Segments U8MinmaxSlice   `json:"segments"`
}

type QueryConfig struct {
	Kv KeyValConfig `json:"kv"`
}

type HeadersConfig struct {
	Kv KeyValConfig `json:"kv"`
}

type Consult struct { // If guard needs to be consulted but is unavaliable => block
	Active             bool   `json:"active"` // False means never consult guard
	RequestsPerMinuete uint16 `json:"rpm"`    // Maximum rpm allows for consulting guard
}

type ReqConfig struct {
	Url     UrlConfig     `json:"url"`
	Qs      QueryConfig   `json:"qs"`
	Headers HeadersConfig `json:"headers"`
}

type WsGate struct {
	Req          ReqConfig `json:"req"`        // Main critiria for blocking/allowing
	ConsultGuard Consult   `json:"consult"`    // If blocked by main critiria, consult guard (if avaliable)
	ForceAllow   bool      `json:"forceAllow"` // Allow no matter what! Overides all blocking.
}
*/
