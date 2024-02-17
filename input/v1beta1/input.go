// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=cel.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This isn't a custom resource, in the sense that we never install its CRD.
// It is a KRM-like object, so we generate a CRD to describe its schema.

// Filters can be used to filter desired composed resources.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Filters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Filters to apply to the desired composed resources produced by previous
	// functions in the pipeline. Each filter matches a desired composed
	// resource by name. If the expression evaluates to true, the composed
	// resource will be included. Desired composed resources that don't match
	// any filter are always included.
	Filters []Filter `json:"filters"`
}

// A Filter can be used to filter a desired composed resource produced by a
// previous function in the pipeline.
type Filter struct {
	// Name of the desired composed resource this filter should match. Supports
	// regular expressions. Only the first matching filter will apply.
	Name string `json:"name"`

	// Expression is a CEL expression. See https://github.com/google/cel-spec.
	Expression string `json:"expression"`
}
