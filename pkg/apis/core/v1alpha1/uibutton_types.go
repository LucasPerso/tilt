/*
Copyright 2020 The Tilt Dev Authors

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

package v1alpha1

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/tilt-dev/tilt-apiserver/pkg/server/builder/resource"
	"github.com/tilt-dev/tilt-apiserver/pkg/server/builder/resource/resourcestrategy"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UIButton
// +k8s:openapi-gen=true
type UIButton struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   UIButtonSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status UIButtonStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// UIButtonList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type UIButtonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []UIButton `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// UIButtonSpec defines the desired state of UIButton
type UIButtonSpec struct {
	// Location associates the button with another component for layout.
	Location UIComponentLocation `json:"location" protobuf:"bytes,1,opt,name=location"`

	// Text to appear on the button itself or as hover text (depending on button location).
	Text string `json:"text" protobuf:"bytes,2,opt,name=text"`

	// IconName is a Material Icon to appear next to button text or on the button itself (depending on button location).
	//
	// Valid values are icon font ligature names from the Material Icons set.
	// See https://fonts.google.com/icons for the full list of available icons.
	//
	// If both IconSVG and IconName are specified, IconSVG will take precedence.
	//
	// +optional
	IconName string `json:"iconName,omitempty" protobuf:"bytes,3,opt,name=iconName"`

	// IconSVG is an SVG to use as the icon to appear next to button text or on the button itself (depending on button
	// location).
	//
	// This should be an <svg> element scaled for a 24x24 viewport.
	//
	// If both IconSVG and IconName are specified, IconSVG will take precedence.
	//
	// +optional
	IconSVG string `json:"iconSVG,omitempty" protobuf:"bytes,4,opt,name=iconSVG"`

	// If true, the button will be rendered, but with an effect indicating it's
	// disabled. It will also be unclickable.
	//
	// +optional
	Disabled bool `json:"disabled,omitempty" protobuf:"varint,5,opt,name=disabled"`
}

// UIComponentLocation specifies where to put a UI component.
type UIComponentLocation struct {
	// ComponentID is the identifier of the parent component to associate this component with.
	//
	// For example, this is a resource name if the ComponentType is Resource.
	ComponentID string `json:"componentID" protobuf:"bytes,1,opt,name=componentID"`
	// ComponentType is the type of the parent component.
	ComponentType ComponentType `json:"componentType" protobuf:"bytes,2,opt,name=componentType,casttype=ComponentType"`
}

type ComponentType string

const (
	ComponentTypeResource ComponentType = "Resource"
	ComponentTypeGlobal   ComponentType = "Global"
)

type UIComponentLocationResource struct {
	ResourceName string `json:"resourceName" protobuf:"bytes,1,opt,name=resourceName"`
}

var _ resource.Object = &UIButton{}
var _ resourcestrategy.Validater = &UIButton{}

func (in *UIButton) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *UIButton) NamespaceScoped() bool {
	return false
}

func (in *UIButton) New() runtime.Object {
	return &UIButton{}
}

func (in *UIButton) NewList() runtime.Object {
	return &UIButtonList{}
}

func (in *UIButton) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "tilt.dev",
		Version:  "v1alpha1",
		Resource: "uibuttons",
	}
}

func (in *UIButton) IsStorageVersion() bool {
	return true
}

func (in *UIButton) Validate(_ context.Context) field.ErrorList {
	var fieldErrors field.ErrorList

	if in.Spec.Text == "" {
		fieldErrors = append(fieldErrors, field.Required(
			field.NewPath("spec.text"), "Button text cannot be empty"))
	}

	locField := field.NewPath("spec.location")
	if in.Spec.Location.ComponentID == "" {
		fieldErrors = append(fieldErrors, field.Required(
			locField.Child("componentID"), "Parent component ID is required"))
	}
	if in.Spec.Location.ComponentType == "" {
		fieldErrors = append(fieldErrors, field.Required(
			locField.Child("componentType"), "Parent component type is required"))
	}

	if in.Spec.IconSVG != "" {
		// do a basic sanity check to catch things like users passing a filename or a <path> directly
		if !strings.Contains(in.Spec.IconSVG, "<svg") {
			fieldErrors = append(fieldErrors, field.Invalid(field.NewPath("spec.iconSVG"), in.Spec.IconSVG,
				"Invalid <svg> element"))
		}
	}

	return fieldErrors
}

var _ resource.ObjectList = &UIButtonList{}

func (in *UIButtonList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

// UIButtonStatus defines the observed state of UIButton
type UIButtonStatus struct {
	// LastClickedAt is the timestamp of the last time the button was clicked.
	//
	// If the button has never clicked before, this will be the zero-value/null.
	LastClickedAt metav1.MicroTime `json:"lastClickedAt,omitempty" protobuf:"bytes,1,opt,name=lastClickedAt"`
}

// UIButton implements ObjectWithStatusSubResource interface.
var _ resource.ObjectWithStatusSubResource = &UIButton{}

func (in *UIButton) GetStatus() resource.StatusSubResource {
	return in.Status
}

// UIButtonStatus{} implements StatusSubResource interface.
var _ resource.StatusSubResource = &UIButtonStatus{}

func (in UIButtonStatus) CopyTo(parent resource.ObjectWithStatusSubResource) {
	parent.(*UIButton).Status = in
}
