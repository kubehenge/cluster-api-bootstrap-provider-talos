// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package controllers

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/siderolabs/crypto/x509"
	"github.com/siderolabs/talos/pkg/machinery/config/container"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapv1beta1 "github.com/kubehenge/cluster-api-bootstrap-provider-talos/api/v1beta1"
)

func TestUserConfigsWithoutClusterSettings(t *testing.T) {
	t.Parallel()

	ca, err := x509.NewSelfSignedCertificateAuthority(x509.Organization("talos"), x509.ECDSA(true))
	require.NoError(t, err)

	machineConfig, err := container.New(&v1alpha1.Config{
		ConfigVersion: "v1alpha1",
		MachineConfig: &v1alpha1.MachineConfig{
			MachineType: "controlplane",
			MachineCA:   &x509.PEMEncodedCertificateAndKey{Crt: ca.CrtPEM, Key: ca.KeyPEM},
		},
	})
	require.NoError(t, err)

	data, err := machineConfig.EncodeString(encoder.WithComments(encoder.CommentsDisabled))
	require.NoError(t, err)

	r := &TalosConfigReconciler{Log: logr.Discard()}
	scope := &TalosConfigScope{
		Config: &bootstrapv1beta1.TalosConfig{
			Spec: bootstrapv1beta1.TalosConfigSpec{GenerateType: "none", Data: data},
		},
	}

	bundle, err := r.userConfigs(t.Context(), scope)
	require.NoError(t, err)

	assert.NotEmpty(t, bundle.BootstrapData)
	assert.NotEmpty(t, bundle.TalosConfig)
}
