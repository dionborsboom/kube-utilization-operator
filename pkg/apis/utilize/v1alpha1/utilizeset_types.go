package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UtilizeSetSpec defines the desired state of UtilizeSet
type UtilizeSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Replicas int32 `json:"replicas"`
	FillToCapacity int32 `json:"filltocapacity"`
	Image string `json:"image"`
	InitScript string `json:"initscript"`
	CPUPerPod int32 `json: "cpuperpod"`
	MemPerPod int32 `json: "memperpod"`
}

// UtilizeSetStatus defines the observed state of UtilizeSet
type UtilizeSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Replicas int32 `json:"replicas"`
	PodNames []string `json:"podNames"`
	Capacity int `json:"capacity"`
	TotalCPU int `json: "totalCPU"`
	TotalMem []string `json: "totalMem"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UtilizeSet is the Schema for the utilizesets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=utilizesets,scope=Namespaced
type UtilizeSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UtilizeSetSpec   `json:"spec,omitempty"`
	Status UtilizeSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UtilizeSetList contains a list of UtilizeSet
type UtilizeSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UtilizeSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UtilizeSet{}, &UtilizeSetList{})
}
