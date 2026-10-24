# Cluster API Bootstrap Provider Talos

The Cluster API Bootstrap Provider Talos (CABPT) is a [Cluster API](https://cluster-api.sigs.k8s.io/) bootstrap provider for [Talos Linux](https://www.talos.dev/).
For every `Machine` or `MachinePool` that references a `TalosConfig`, it generates the Talos machine configuration, or passes through one you supply, and stores it as the bootstrap data secret Cluster API hands to the infrastructure provider.
It also keeps the cluster-wide Talos secrets and a `talosconfig` for the Talos API of the cluster.

It is used together with a control plane provider such as the [Cluster API Control Plane Provider Talos](https://github.com/kubehenge/cluster-api-control-plane-provider-talos).

This fork is maintained primarily for [Kubehenge](https://kubehenge.dev), and it is open to anyone: issues and pull requests are welcome.
Report bugs and ask questions in [this repository's issues](https://github.com/kubehenge/cluster-api-bootstrap-provider-talos/issues), and security problems privately, as [SECURITY.md](.github/SECURITY.md) describes; [CONTRIBUTING.md](.github/CONTRIBUTING.md) covers pull requests.

## Compatibility

| Component | Supported |
| --- | --- |
| Cluster API contract | v1beta2 |
| Cluster API | v1.14, built and tested against v1.14.3 |
| Talos Linux | v1.9 to v1.14; configuration is generated for v1.14 unless `talosVersion` names an earlier release |
| Kubernetes | the releases the Talos release supports, listed below; v1.37.1 when the `Machine` names no version |

| Talos Linux | Kubernetes |
| --- | --- |
| v1.14 | v1.32 to v1.37 |
| v1.13 | v1.31 to v1.36 |
| v1.12 | v1.30 to v1.35 |
| v1.11 | v1.29 to v1.34 |
| v1.10 | v1.28 to v1.33 |
| v1.9 | v1.27 to v1.32 |

The integration tests run against Talos v1.14 and generate configuration for v1.13 and v1.14.

## Installing

clusterctl knows a provider called `talos` that points at the Sidero Labs releases.
To install this fork instead, add it to `~/.config/cluster-api/clusterctl.yaml` (or `~/.cluster-api/clusterctl.yaml`):

```yaml
providers:
  - name: talos
    url: https://github.com/kubehenge/cluster-api-bootstrap-provider-talos/releases/latest/bootstrap-components.yaml
    type: BootstrapProvider
```

Then initialise the management cluster with it:

```shell
clusterctl init --bootstrap talos --control-plane talos --infrastructure <infrastructure provider>
```

clusterctl never picks a prerelease on its own; to install one, name its version, as in `--bootstrap talos:<version>`.

## Usage

Control plane machines get their `TalosConfig` from the control plane provider, worker machines from a `TalosConfigTemplate`:

```yaml
apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
kind: TalosConfigTemplate
metadata:
  name: workers
spec:
  template:
    metadata:
      labels:
        example.com/pool: workers
    spec:
      generateType: worker
      talosVersion: v1.14
      strategicPatches:
        - |
          apiVersion: v1alpha1
          kind: UnattendedInstallConfig
          provisioning:
            diskSelector:
              match: disk.dev_path == "/dev/sda"
```

- `generateType`: the machine type to generate configuration for (`controlplane`, `worker`, or the deprecated `init`), or `none` to use the configuration in `data` as it is.
- `talosVersion`: the Talos release to generate configuration for, such as `v1.14`. Set it explicitly, so that a newer provider does not move existing machines to a newer configuration format.
- `strategicPatches` and `configPatches`: strategic merge patches and RFC 6902 JSON patches applied to the configuration.
- `hostname.source`: `MachineName` or `InfrastructureName` sets the hostname from the `Machine` or its infrastructure machine.
- `spec.template.metadata`: labels and annotations Cluster API copies onto the `TalosConfig` objects it creates from the template. Only `spec.template.spec` is immutable.

The provider stores the `talosconfig` of a cluster in the `<cluster>-talosconfig` secret:

```shell
kubectl get secret --namespace <namespace> <cluster>-talosconfig -o jsonpath='{.data.talosconfig}' | base64 -d > talosconfig
```

## Releases

A release is built by the release workflow from its version tag and carries `bootstrap-components.yaml`, `metadata.yaml` and `SHA256SUMS`.

- The controller image is `ghcr.io/kubehenge/cluster-api-bootstrap-provider-talos`, and `bootstrap-components.yaml` names it by digest, so what a release installs cannot change after it is published.
- The image is signed with keyless [cosign](https://docs.sigstore.dev/) and carries a build provenance attestation; the release files carry one too.
- `SHA256SUMS` holds the checksums of the release files.

To check the image a release names before installing it:

```shell
cosign verify ghcr.io/kubehenge/cluster-api-bootstrap-provider-talos@sha256:<digest> \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github\.com/kubehenge/cluster-api-bootstrap-provider-talos/\.github/workflows/release\.yaml@refs/tags/v'
gh attestation verify oci://ghcr.io/kubehenge/cluster-api-bootstrap-provider-talos@sha256:<digest> \
  --repo kubehenge/cluster-api-bootstrap-provider-talos
```

And the downloaded release files:

```shell
sha256sum -c SHA256SUMS
gh attestation verify bootstrap-components.yaml --repo kubehenge/cluster-api-bootstrap-provider-talos
```

## Building

The Go module is `github.com/kubehenge/cluster-api-bootstrap-provider-talos`.
Everything builds natively with Go and the tools the Makefile installs into `_out/bin`; `make help` lists the targets.

```shell
make generate manifests  # deepcopy and conversion code, CRDs, webhooks and RBAC
make lint unit           # golangci-lint and the unit tests under the race detector
make image KO_FLAGS=--push=false TAG=dev
```

The integration tests need Docker on linux/amd64: `make env-up integration` creates a Talos cluster with talosctl and runs them against it.
Every image and manifest that environment pulls is pinned by digest or SHA-256 ([internal/integration](internal/integration/README.md)).

## Credits

This provider was created by [Sidero Labs](https://www.siderolabs.com/), who keep the original at [siderolabs/cluster-api-bootstrap-provider-talos](https://github.com/siderolabs/cluster-api-bootstrap-provider-talos).
A few fixes are pulled in from the community's forks.
[CHANGELOG.md](CHANGELOG.md) lists the changes of this fork's releases and links the history of the Sidero Labs releases.

## Licence

[Mozilla Public License 2.0](LICENSE). `internal/configowner/configowner.go` is copied from Cluster API and stays under the Apache License 2.0 named in it.
