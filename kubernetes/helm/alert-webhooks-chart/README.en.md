# Helm Installation Guide

This guide installs, upgrades, verifies, and rolls back Alert Webhooks from its
public Helm repository. Source-based validation commands run from the repository
root unless a section states otherwise.

## Prerequisites

- A reachable Kubernetes cluster and a kubeconfig with access to the target
  namespace.
- Helm and kubectl.

On macOS, install the required tools with Homebrew:

```bash
brew install helm kubectl
helm version
kubectl version --client
kubectl config current-context
```

For other operating systems, see the official
[Helm installation guide](https://helm.sh/docs/intro/install/) and
[kubectl installation guide](https://kubernetes.io/docs/tasks/tools/).

Before making changes, confirm that `kubectl config current-context` points to
the intended cluster.

## Security configuration

The chart does not provide default webhook, metric, or notification-service
credentials. Supply the following values through a private values file:

- `secrets.metric.user`
- `secrets.metric.password`
- `secrets.webhooks.user`
- `secrets.webhooks.password`

When `config.telegram.enable`, `config.slack.enable`, or
`config.discord.enable` is `true`, provide the corresponding
`secrets.<service>.token`. Trace authentication remains optional.

Do not commit real credentials to the repository. Avoid `--set` for secrets
because it can leave values in shell history or CI logs. Store the private
values file outside the repository and restrict its file permissions.

The following minimum example disables all notification services. Add tokens
through the private values file only when enabling a service:

```yaml
ingress:
  enabled: false

config:
  telegram:
    enable: false
  slack:
    enable: false
  discord:
    enable: false

secrets:
  metric:
    user: "REPLACE_METRIC_USER"
    password: "REPLACE_METRIC_PASSWORD"
  webhooks:
    user: "REPLACE_WEBHOOKS_USER"
    password: "REPLACE_WEBHOOKS_PASSWORD"
```

## Version contract

- Chart versions and application versions are managed independently.
- A `chart-v<chart version>` tag publishes only the matching immutable package.
- This release uses chart version `1.0.1`, application version `v2.0.8`, and
  container image tag `v2.0.8` by default.
- Published chart versions are not overwritten, and the repository does not
  provide a mutable `latest` chart.

## Install from the public repository

Add the public repository and select an explicit chart version:

```bash
helm repo add alert-webhooks https://vincent119.github.io/alert-webhooks
helm repo update
helm show chart alert-webhooks/alert-webhooks-chart --version 1.0.1
```

Set the path to a private values file outside the repository:

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
```

Render the public chart without modifying the cluster:

```bash
helm template alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.1 \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

Review the rendered output for placeholders, unexpected image references, and
secrets before continuing.

## Install or upgrade

```bash
helm upgrade --install alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.1 \
  --namespace alert-webhooks \
  --create-namespace \
  --values "${PRIVATE_VALUES_FILE}" \
  --atomic \
  --wait \
  --timeout 5m
```

The `--atomic` option rolls back a failed upgrade automatically. Stop and
inspect the release if the command does not complete successfully.

## Validate from source

Set the chart directory and use the same private values file:

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
CHART_DIR=kubernetes/helm/alert-webhooks-chart
```

Validate and render the chart locally before modifying the cluster:

```bash
helm lint "${CHART_DIR}" --values "${PRIVATE_VALUES_FILE}"
helm template alert-webhooks "${CHART_DIR}" \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

Review the rendered output for placeholders, unexpected image references, and
secrets before continuing.

## Verify the deployment

```bash
helm status alert-webhooks --namespace alert-webhooks
kubectl get pods,service,ingress --namespace alert-webhooks
kubectl rollout status deployment \
  --namespace alert-webhooks \
  --selector app.kubernetes.io/instance=alert-webhooks \
  --timeout 5m
```

Do not continue with notification testing until the rollout completes and the
Pods are ready.

## Roll back or uninstall

List release revisions before selecting the revision to restore:

```bash
helm history alert-webhooks --namespace alert-webhooks
REVISION_NUMBER=1
helm rollback alert-webhooks "${REVISION_NUMBER}" \
  --namespace alert-webhooks \
  --wait \
  --timeout 5m
```

Uninstall the release with:

```bash
helm uninstall alert-webhooks --namespace alert-webhooks
```

`helm uninstall` does not delete the namespace automatically. Before deleting
the namespace, confirm that it contains no unrelated workloads or retained
resources.

## Chart values and maintenance

The chart source and this guide are in the same directory. Review
[`values.yaml`](./values.yaml) for the complete values contract.

The generated chart README is maintained from `README.md.gotmpl` with the
repository command:

```bash
make helm-docs
```

Do not edit the generated values table manually.

## Related documentation

- [Docker deployment](../../../documentation/en/docker-deployment.md)
- [Kubernetes environment variables](../../../documentation/en/kubernetes-env-vars.md)
- [Kustomize installation](../../kustomize/README.en.md)
