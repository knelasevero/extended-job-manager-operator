package v1

import (
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExtendedJobManager is the Schema for the extendedjobmanager API
// +k8s:openapi-gen=true
// +genclient
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type ExtendedJobManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec holds user settable values for configuration
	// +required
	Spec ExtendedJobManagerSpec `json:"spec"`
	// status holds observed values from the cluster. They may not be overridden.
	// +optional
	Status ExtendedJobManagerStatus `json:"status"`
}

// ExtendedJobManagerSpec defines the desired state of ExtendedJobManager
type ExtendedJobManagerSpec struct {
	operatorv1.OperatorSpec `json:",inline"`

	// ExtendedJobManagerConfig allows configuring the Extended Job Manager.
	ExtendedJobManagerConfig string `json:"extendedJobManagerConfig"`

	// ExtendedJobManagerImage sets the container image url to be pulled
	ExtendedJobManagerImage string `json:"extendedJobManagerImage"`
}

// ExtendedJobManagerStatus defines the observed state of ExtendedJobManager
type ExtendedJobManagerStatus struct {
	operatorv1.OperatorStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExtendedJobManagerList contains a list of ExtendedJobManager
type ExtendedJobManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExtendedJobManager `json:"items"`
}
