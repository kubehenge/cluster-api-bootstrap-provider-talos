// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package controllers

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	bootstrapv1beta1 "github.com/kubehenge/cluster-api-bootstrap-provider-talos/api/v1beta1"
)

func TestReconcileSummarisesReadyCondition(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, bootstrapv1beta1.AddToScheme(scheme))
	require.NoError(t, capiv1.AddToScheme(scheme))

	config := &bootstrapv1beta1.TalosConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config", Namespace: "default"},
		Spec:       bootstrapv1beta1.TalosConfigSpec{GenerateType: "worker"},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(config).WithStatusSubresource(config).Build()
	r := &TalosConfigReconciler{Client: c, Log: logr.Discard(), Scheme: scheme}

	key := types.NamespacedName{Namespace: config.Namespace, Name: config.Name}

	// Without an owner the reconcile stops early, before any bootstrap data is reported.
	_, err := r.Reconcile(t.Context(), ctrl.Request{NamespacedName: key})
	require.NoError(t, err)

	var reconciled bootstrapv1beta1.TalosConfig
	require.NoError(t, c.Get(t.Context(), key, &reconciled))

	ready := conditions.Get(&reconciled, bootstrapv1beta1.TalosConfigReadyCondition)
	require.NotNil(t, ready)
	assert.Equal(t, metav1.ConditionUnknown, ready.Status)
	assert.Equal(t, bootstrapv1beta1.TalosConfigReadyUnknownReason, ready.Reason)
}
