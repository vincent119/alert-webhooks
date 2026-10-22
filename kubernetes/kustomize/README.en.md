# Kustomize Installation Guide

This guide deploys Alert Webhooks with the Kustomize overlay in this
repository. Run all commands from the repository root unless a section states
otherwise.

## Prerequisites

- A reachable Kubernetes cluster and a kubeconfig with access to the target
  namespace.
- kubectl. Kubernetes 1.14 and later include Kustomize support through
  `kubectl kustomize` and `kubectl apply -k`.
- Optional: the standalone Kustomize CLI.

On macOS, install the tools with Homebrew:

```bash
brew install kubectl
brew install kustomize
kubectl version --client
kubectl kustomize --help
```

For other operating systems, see the official
[kubectl installation guide](https://kubernetes.io/docs/tasks/tools/) and
[Kustomize installation guide](https://kubectl.docs.kubernetes.io/installation/kustomize/).

## Pre-deployment checks

### Confirm the Kubernetes context

```bash
kubectl config current-context
kubectl auth can-i create deployments --namespace alert-webhooks
```

Stop if the context or permissions do not match the intended target.

### Review the image and Ingress settings

Review [`kustomization.yaml`](./kustomization.yaml) before deployment:

- `images[].newName` and `images[].newTag` must identify an image that the
  cluster can pull.
- Replace the Ingress certificate ARN and host with values for the target
  environment.
- Adjust or disable the Ingress when the cluster does not use AWS Load
  Balancer Controller.

Do not write private registry credentials, certificates, or internal domains
to the public repository. Manage environment-specific values through a
controlled overlay or deployment process.

### Create the required Secret

The Deployment requires `alert-webhooks-secrets` in the `alert-webhooks`
namespace. Create the namespace first, then create the Secret through the
approved secret-management process.

The following command contains placeholders only. Replace them from a secure
source before running it:

```bash
kubectl create namespace alert-webhooks \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic alert-webhooks-secrets \
  --namespace alert-webhooks \
  --from-literal=metric-user="REPLACE_METRIC_USER" \
  --from-literal=metric-password="REPLACE_METRIC_PASSWORD" \
  --from-literal=webhooks-user="REPLACE_WEBHOOKS_USER" \
  --from-literal=webhooks-password="REPLACE_WEBHOOKS_PASSWORD" \
  --from-literal=telegram-token="REPLACE_TELEGRAM_TOKEN" \
  --from-literal=slack-token="REPLACE_SLACK_TOKEN" \
  --from-literal=discord-token="REPLACE_DISCORD_TOKEN" \
  --from-literal=trace-auth-user="REPLACE_TRACE_AUTH_USER" \
  --from-literal=trace-auth-passwd="REPLACE_TRACE_AUTH_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -
```

Do not write real Secret values to `kustomization.yaml`, shell history, CI
logs, or the repository. Production environments should use the existing
secret-management solution.

## Render and preview changes

Render the manifests locally and check for placeholders, incorrect images, and
unexpected resources:

```bash
kubectl kustomize kubernetes/kustomize \
  > /tmp/alert-webhooks-kustomize.yaml
```

When cluster access is available, preview changes before applying them:

```bash
kubectl diff -k kubernetes/kustomize
```

`kubectl diff` can return a non-zero exit code when differences exist. Review
its output before deciding whether to deploy.

## Deploy

```bash
kubectl apply -k kubernetes/kustomize
```

## Verify the deployment

```bash
kubectl get pods --namespace alert-webhooks
kubectl get service --namespace alert-webhooks
kubectl get deployment --namespace alert-webhooks
kubectl get ingress --namespace alert-webhooks
kubectl rollout status deployment/prod-alert-webhooks \
  --namespace alert-webhooks \
  --timeout 5m
kubectl logs --namespace alert-webhooks --selector app=alert-webhooks
```

Stop and investigate if the rollout does not complete or the Pods are not
ready.

## Update the deployment

After updating the image tag in [`kustomization.yaml`](./kustomization.yaml) or
a controlled overlay, render and preview the changes before applying them:

```bash
kubectl kustomize kubernetes/kustomize \
  > /tmp/alert-webhooks-kustomize.yaml
kubectl diff -k kubernetes/kustomize
kubectl apply -k kubernetes/kustomize
kubectl rollout status deployment/prod-alert-webhooks \
  --namespace alert-webhooks \
  --timeout 5m
```

## Troubleshooting

```bash
kubectl get events --namespace alert-webhooks --sort-by='.lastTimestamp'
kubectl describe deployment prod-alert-webhooks --namespace alert-webhooks
kubectl describe pod --namespace alert-webhooks --selector app=alert-webhooks
kubectl get configmap prod-alert-webhooks-config \
  --namespace alert-webhooks -o yaml
kubectl get secret alert-webhooks-secrets --namespace alert-webhooks
```

Do not use options that print Secret data, and do not paste Secret output into
an issue, chat, or CI log.

## Clean up resources

Preview the resources that would be deleted:

```bash
kubectl delete -k kubernetes/kustomize --dry-run=client -o name
```

After reviewing the list, delete the Kustomize-managed resources:

```bash
kubectl delete -k kubernetes/kustomize
```

The externally created `alert-webhooks-secrets` Secret is not managed by
Kustomize and must be deleted separately when appropriate. Before deleting the
namespace, confirm that it contains no unrelated workloads.

## Repository layout

```text
kubernetes/kustomize/
├── README.md
├── kustomization.yaml
├── base/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
└── configs/
    ├── config.production.yaml
    ├── alert_config.yaml
    └── alert_config.minimal.yaml
```

## Related documentation

- [Docker deployment](../../documentation/en/docker-deployment.md)
- [Kubernetes environment variables](../../documentation/en/kubernetes-env-vars.md)
- [Helm installation](../helm/alert-webhooks-chart/README.md)
