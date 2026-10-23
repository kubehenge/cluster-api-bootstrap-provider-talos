// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1beta1_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	bootstrapv1beta1 "github.com/kubehenge/cluster-api-bootstrap-provider-talos/api/v1beta1"
)

func TestTemplateDeepCopyCopiesTheMetadata(t *testing.T) {
	t.Parallel()

	template := &bootstrapv1beta1.TalosConfigTemplate{
		Spec: bootstrapv1beta1.TalosConfigTemplateSpec{
			Template: bootstrapv1beta1.TalosConfigTemplateResource{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels:      map[string]string{"example.com/pool": "workers"},
					Annotations: map[string]string{"example.com/owner": "team"},
				},
			},
		},
	}

	copied := template.DeepCopy()
	copied.Spec.Template.ObjectMeta.Labels["example.com/pool"] = "changed"
	copied.Spec.Template.ObjectMeta.Annotations["example.com/owner"] = "changed"

	assert.Equal(t, "workers", template.Spec.Template.ObjectMeta.Labels["example.com/pool"])
	assert.Equal(t, "team", template.Spec.Template.ObjectMeta.Annotations["example.com/owner"])
}
