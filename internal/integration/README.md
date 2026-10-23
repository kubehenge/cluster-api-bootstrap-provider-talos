# Running integration tests

The tests run against a Talos cluster in Docker on linux/amd64, as CI runs them:

```sh
make env-up integration
make env-down
```

`make env-up` creates the cluster with the pinned talosctl and writes its kubeconfig to `_out/env`.
`make integration` installs Cluster API and its Docker infrastructure provider into that cluster with
clusterctl, at the release go.mod requires, installs the CRDs from `config`, and runs the controller
and its webhooks in the test process.

Tests clean up after the run, so they can be run repeatedly against the cluster;
`INTEGRATION_SKIP_CLEANUP=1` keeps what they create.

## Pinned images and manifests

Nothing the environment pulls is resolved by tag. `env/` holds every pin, and Renovate keeps each
one current:

- `talos-patch.yaml` and `talos-patch-controlplanes.yaml` are the machine configuration patches
  `make env-up` gives talosctl. They name the kubelet, the API server, controller manager,
  scheduler, kube-proxy, etcd, CoreDNS and the pod sandbox image by digest. Talos offers no setting
  for its Flannel image, so its Flannel is replaced by the same manifests with the image named by
  digest.
- `components.txt` lists the Cluster API and cert-manager manifests with their SHA-256 and the
  images they name with their digests. `hack/integration-components.sh` fetches and checks the
  manifests, rewrites their images to the digests and writes a clusterctl configuration that
  installs from that local repository instead of GitHub.

After the tests, `hack/check-images.sh` lists the images on every node and fails when one was pulled
by tag or by a digest nothing pins.
