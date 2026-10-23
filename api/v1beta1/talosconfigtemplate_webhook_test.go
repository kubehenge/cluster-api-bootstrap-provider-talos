// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1beta1_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	bootstrapv1beta1 "github.com/siderolabs/cluster-api-bootstrap-provider-talos/api/v1beta1"
)

func updateContext() context.Context {
	return admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{Operation: admissionv1.Update},
	})
}

// dryRunContext carries the dry-run admission request the topology controller issues while
// computing the desired state of a Cluster.
func dryRunContext() context.Context {
	return admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
			DryRun:    ptr.To(true),
		},
	})
}

func configTemplate(annotations map[string]string, template bootstrapv1beta1.TalosConfigTemplateResource) *bootstrapv1beta1.TalosConfigTemplate {
	return &bootstrapv1beta1.TalosConfigTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: annotations,
		},
		Spec: bootstrapv1beta1.TalosConfigTemplateSpec{Template: template},
	}
}

func templateResource(version string) bootstrapv1beta1.TalosConfigTemplateResource {
	return bootstrapv1beta1.TalosConfigTemplateResource{
		Spec: bootstrapv1beta1.TalosConfigSpec{GenerateType: "worker", TalosVersion: version},
	}
}

func invalidMetadata() clusterv1.ObjectMeta {
	return clusterv1.ObjectMeta{Labels: map[string]string{"not a valid key": "workers"}}
}

func TestTemplateValidateUpdate_TemplateSpecIsImmutable(t *testing.T) {
	t.Parallel()

	old := configTemplate(nil, templateResource("v1.12"))
	updated := configTemplate(nil, templateResource("v1.13"))

	_, err := updated.ValidateUpdate(updateContext(), old, updated)

	require.Error(t, err)
	assert.True(t, apierrors.IsInvalid(err))
	assert.Contains(t, err.Error(), "spec.template.spec is immutable")
}

// Template metadata is only copied onto bootstrap configs generated later, so changing it cannot
// invalidate machines that already exist.
func TestTemplateValidateUpdate_TemplateMetadataIsMutable(t *testing.T) {
	t.Parallel()

	old := configTemplate(nil, templateResource("v1.13"))

	updatedTemplate := templateResource("v1.13")
	updatedTemplate.ObjectMeta = clusterv1.ObjectMeta{
		Labels:      map[string]string{"example.com/pool": "workers"},
		Annotations: map[string]string{"example.com/note": "added later"},
	}
	updated := configTemplate(nil, updatedTemplate)

	_, err := updated.ValidateUpdate(updateContext(), old, updated)

	require.NoError(t, err)
}

// The topology controller dry-runs the template it would write; that probe is admitted even though
// it changes the template spec.
func TestTemplateValidateUpdate_TemplateSpecChangeAdmittedDuringTopologyDryRun(t *testing.T) {
	t.Parallel()

	annotations := map[string]string{clusterv1.TopologyDryRunAnnotation: ""}

	old := configTemplate(nil, templateResource("v1.12"))
	updated := configTemplate(annotations, templateResource("v1.13"))

	_, err := updated.ValidateUpdate(dryRunContext(), old, updated)

	require.NoError(t, err)
}

// A dry-run from any other client (kubectl apply --dry-run=server, say) stays subject to
// immutability, because it lacks the topology annotation.
func TestTemplateValidateUpdate_TemplateSpecChangeRejectedDuringPlainDryRun(t *testing.T) {
	t.Parallel()

	old := configTemplate(nil, templateResource("v1.12"))
	updated := configTemplate(nil, templateResource("v1.13"))

	_, err := updated.ValidateUpdate(dryRunContext(), old, updated)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "spec.template.spec is immutable")
}

func TestTemplateValidateCreate_RejectsInvalidTemplateMetadata(t *testing.T) {
	t.Parallel()

	template := templateResource("v1.13")
	template.ObjectMeta = invalidMetadata()

	_, err := (&bootstrapv1beta1.TalosConfigTemplate{}).ValidateCreate(context.Background(), configTemplate(nil, template))

	require.Error(t, err)
	assert.True(t, apierrors.IsInvalid(err))
	assert.Contains(t, err.Error(), "spec.template.metadata.labels")
}

func TestTemplateValidateCreate_AdmitsValidTemplateMetadata(t *testing.T) {
	t.Parallel()

	template := templateResource("v1.13")
	template.ObjectMeta = clusterv1.ObjectMeta{Labels: map[string]string{"example.com/pool": "workers"}}

	_, err := (&bootstrapv1beta1.TalosConfigTemplate{}).ValidateCreate(context.Background(), configTemplate(nil, template))

	require.NoError(t, err)
}

func TestTemplateValidateUpdate_RejectsInvalidTemplateMetadata(t *testing.T) {
	t.Parallel()

	old := configTemplate(nil, templateResource("v1.13"))

	updatedTemplate := templateResource("v1.13")
	updatedTemplate.ObjectMeta = invalidMetadata()
	updated := configTemplate(nil, updatedTemplate)

	_, err := updated.ValidateUpdate(updateContext(), old, updated)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "spec.template.metadata.labels")
}
