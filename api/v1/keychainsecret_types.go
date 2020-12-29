/*
Copyright (c) 2020 Facebook, Inc. and its affiliates.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeychainSecretSpec defines the desired state of KeychainSecret
type KeychainSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of KeychainSecret. Edit KeychainSecret_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// KeychainSecretStatus defines the observed state of KeychainSecret
type KeychainSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// KeychainSecret is the Schema for the keychainsecrets API
type KeychainSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeychainSecretSpec   `json:"spec,omitempty"`
	Status KeychainSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeychainSecretList contains a list of KeychainSecret
type KeychainSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeychainSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeychainSecret{}, &KeychainSecretList{})
}
