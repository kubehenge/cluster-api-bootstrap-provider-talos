# Changelog

The changes of each release of this fork. The release pages on GitHub list every commit.

## v0.7.0-alpha.5

- The Go module is now `github.com/kubehenge/cluster-api-bootstrap-provider-talos`; code that imports the API types has to change its import paths.
- The `cabpt-controller-manager-metrics-service` service now forwards to the manager's metrics port.
- The manifests are rendered without deprecated kustomize fields; what `bootstrap-components.yaml` installs is otherwise unchanged.

## v0.7.0-alpha.4

The first release of this fork, built on Sidero Labs' v0.7.0-alpha.3.

- Built against Cluster API v1.14.3 and the Cluster API v1beta2 contract.
- `TalosConfigTemplate` gains `spec.template.metadata`: labels and annotations Cluster API copies onto the `TalosConfig` objects it creates. Only `spec.template.spec` is immutable.
- A Talos 1.14 machine configuration with a machine CA and no cluster section no longer makes the controller panic.
- The v1alpha3 conversions are registered with the scheme, so converting through the scheme no longer fails with "unknown conversion".
- The `Ready` summary condition is set.
- The controller may update the finalizers of `TalosConfig` and `Cluster` objects, which the secrets it creates need on API servers that enforce owner reference permissions.
- The configuration patches, which can carry secrets, are no longer logged at info level on every reconcile.
- The controller image is `ghcr.io/kubehenge/cluster-api-bootstrap-provider-talos`, built with ko, signed with keyless cosign and carrying build provenance; `bootstrap-components.yaml` names it by digest.

## Before the fork

Sidero Labs' releases up to v0.7.0-alpha.3 are described in [their changelog](https://github.com/siderolabs/cluster-api-bootstrap-provider-talos/blob/v0.7.0-alpha.3/CHANGELOG.md).
