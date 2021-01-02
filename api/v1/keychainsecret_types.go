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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file

// For kubebuilder marker syntax see:
// https://book.kubebuilder.io/reference/markers/crd-validation.html

// KeychainSecretSpec defines the desired state of KeychainSecret
type KeychainSecretSpec struct {
	// Name is the name of the Keychain secret.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=150
	// +kubebuilder:validation:Pattern="^[A-Z0-9_]+$"
	Name string `json:"name,required"`
	// Group is the name of the Keychain group the secret exist in. It is optional as not all secrets exit in a group.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=150
	// +kubebuilder:validation:Pattern="^[A-Z0-9_]+$"
	// +optional
	Group string `json:"group,omitempty"`
	// TTL is how often this secret should be updated (for rotation purposes). It is a golang Duration, and we use a
	// regex to validate it. Note that only seconds (s), minutes (m), or hours (h) are allowed because durations
	// involving days or years may be ambiguous due to differences in locales. See https://github.com/golang/go/issues/17767
	// for the "official" rational...
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="^[0-9]+[smh]$"
	// +kubebuilder:default="24h"
	// +optional
	TTL string `json:"ttl,omitempty"`
}

// KeychainSecretStatus defines the observed state of KeychainSecret
type KeychainSecretStatus struct {
	// SecretRef is a reference to the Secret this KeychainSecret created and maintains.
	// +optional
	SecretRef corev1.SecretReference `json:"secretRef,omitempty"`
	// LastUpdate is the time we updated this secret.
	// It is a fixed, portable, seriallized version of the golang type https://golang.org/pkg/time/#Time
	// +optional
	LastUpdate metav1.Time `json:"lastUpdate,omitempty" protobuf:"bytes,3,opt,name=lastUpdate"`
	// Message is human-readable string indicating details about the last update.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
	// Reason is a brief CamelCase string that describes any failure and is meant
	// for machine parsing and tidy display in the CLI.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
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
