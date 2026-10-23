/*
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

package configowner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	capiv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/siderolabs/cluster-api-bootstrap-provider-talos/internal/configowner"
)

// owned returns an object owned by the given Cluster API kind.
func owned(kind, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config", Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: capiv1.GroupVersion.String(), Kind: kind, Name: name, UID: "uid"}},
		},
	}
}

func fakeClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, capiv1.AddToScheme(scheme))

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func TestGetConfigOwnerReturnsTheOwningMachine(t *testing.T) {
	t.Parallel()

	machine := &capiv1.Machine{
		ObjectMeta: metav1.ObjectMeta{Name: "machine-1", Namespace: "default"},
		Spec: capiv1.MachineSpec{
			ClusterName: "test",
			Bootstrap:   capiv1.Bootstrap{DataSecretName: ptr.To("machine-1-data")},
		},
	}

	owner, err := configowner.GetConfigOwner(context.Background(), fakeClient(t, machine), owned("Machine", "machine-1"))
	require.NoError(t, err)
	require.NotNil(t, owner)

	assert.Equal(t, "machine-1", owner.GetName())
	assert.Equal(t, "default", owner.GetNamespace())
	assert.Equal(t, "Machine", owner.GetKind())
	assert.False(t, owner.IsMachinePool())
	assert.Equal(t, "test", owner.ClusterName())
	require.NotNil(t, owner.DataSecretName())
	assert.Equal(t, "machine-1-data", *owner.DataSecretName())
}

func TestGetConfigOwnerReturnsTheOwningMachinePool(t *testing.T) {
	t.Parallel()

	pool := &capiv1.MachinePool{
		ObjectMeta: metav1.ObjectMeta{Name: "pool-1", Namespace: "default"},
		Spec:       capiv1.MachinePoolSpec{ClusterName: "test"},
	}

	owner, err := configowner.GetConfigOwner(context.Background(), fakeClient(t, pool), owned("MachinePool", "pool-1"))
	require.NoError(t, err)
	require.NotNil(t, owner)

	assert.True(t, owner.IsMachinePool())
	assert.Equal(t, "test", owner.ClusterName())
	assert.Nil(t, owner.DataSecretName(), "a MachinePool carries no bootstrap data secret name")
}

func TestGetConfigOwnerIgnoresOtherOwners(t *testing.T) {
	t.Parallel()

	owner, err := configowner.GetConfigOwner(context.Background(), fakeClient(t), owned("Deployment", "app"))
	require.NoError(t, err)
	assert.Nil(t, owner)
}
