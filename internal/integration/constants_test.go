// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package integration

// TalosVersion is the Talos release of the cluster the tests run against; make integration sets it.
var TalosVersion string

// CAPIVersion is the Cluster API release the tests install; make integration sets it.
var CAPIVersion string

// Components is the rendered provider components file the CRDs are read from; make integration sets it.
var Components string

// ClusterctlConfig is the clusterctl configuration naming the local repository Cluster API and
// cert-manager are installed from, with their manifests checked and their images pinned by digest;
// make integration sets it.
var ClusterctlConfig string
