/*
Copyright 2021.

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

type CalibreJobSpec struct {
	Schedule string `json:"schedule,omitempty"`
	Command  string `json:"command,omitempty"`
}

type CalibreJobStatus struct {
	Phase string `json:"phase,omitempty"`
}

const (
	PhasePending = "PENDING"
	PhaseRunning = "RUNNING"
	PhaseDone    = "DONE"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.phase", name=Phase, type=string

type CalibreJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CalibreJobSpec   `json:"spec,omitempty"`
	Status CalibreJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type CalibreJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CalibreJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CalibreJob{}, &CalibreJobList{})
}