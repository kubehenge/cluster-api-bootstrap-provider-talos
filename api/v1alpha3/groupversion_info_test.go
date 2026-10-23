// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package v1alpha3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	bootstrapv1alpha3 "github.com/siderolabs/cluster-api-bootstrap-provider-talos/api/v1alpha3"
	bootstrapv1beta1 "github.com/siderolabs/cluster-api-bootstrap-provider-talos/api/v1beta1"
)

func TestAddToSchemeRegistersEveryKind(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, bootstrapv1alpha3.AddToScheme(scheme))

	for _, kind := range []string{"TalosConfig", "TalosConfigList", "TalosConfigTemplate", "TalosConfigTemplateList"} {
		assert.True(t, scheme.Recognizes(bootstrapv1alpha3.GroupVersion.WithKind(kind)), kind)
	}
}

// The scheme converts through the hand-written status conversion, which moves the v1beta1
// contract failure fields into the deprecated block of the hub.
func TestAddToSchemeRegistersTheConversions(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, bootstrapv1alpha3.AddToScheme(scheme))
	require.NoError(t, bootstrapv1beta1.AddToScheme(scheme))

	spoke := &bootstrapv1alpha3.TalosConfig{
		Spec:   bootstrapv1alpha3.TalosConfigSpec{GenerateType: "worker"},
		Status: bootstrapv1alpha3.TalosConfigStatus{FailureReason: "reason", FailureMessage: "message"},
	}

	hub := &bootstrapv1beta1.TalosConfig{}
	require.NoError(t, scheme.Convert(spoke, hub, nil))

	assert.Equal(t, "worker", hub.Spec.GenerateType)
	require.NotNil(t, hub.Status.Deprecated)
	require.NotNil(t, hub.Status.Deprecated.V1Beta1)
	assert.Equal(t, "reason", hub.Status.Deprecated.V1Beta1.FailureReason)
	assert.Equal(t, "message", hub.Status.Deprecated.V1Beta1.FailureMessage)
}
