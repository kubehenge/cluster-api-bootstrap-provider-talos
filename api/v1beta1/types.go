// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1beta1

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// TalosConfigTemplateResource defines the Template structure
type TalosConfigTemplateResource struct {
	// metadata holds the labels and annotations Cluster API copies onto the TalosConfigs it
	// generates from this template.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	Spec TalosConfigSpec `json:"spec,omitempty"`
}

// nb: we use apiextensions.JSON for the value below b/c we can't use interface{} with controller-gen.
// found this workaround here: https://github.com/kubernetes-sigs/controller-tools/pull/126#issuecomment-630769075

type ConfigPatches struct {
	Op    string             `json:"op"`
	Path  string             `json:"path"`
	Value apiextensions.JSON `json:"value,omitempty"`
}
