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
