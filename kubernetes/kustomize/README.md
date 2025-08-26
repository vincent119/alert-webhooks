# Kustomize Setup

## Deploy Application

### Method 1: Direct Deployment with Kustomize

```bash
# Preview generated configuration
kubectl kustomize .

```

```bash
# Deploy directly
kubectl apply -k .

```

### Method 2: Step-by-Step Deployment

```bash
# 1. Create namespace
kubectl create namespace alert-webhooks
```

```bash
# 2. Create Secret
kubectl create secret generic alert-webhooks-secrets \
  -n alert-webhooks \
  --from-literal=metric-user="admin" \
  --from-literal=metric-password="admin" \
  --from-literal=webhooks-user="admin" \
  --from-literal=webhooks-password="admin" \
  --from-literal=telegram-token="your-telegram-token" \
  --from-literal=slack-token="your-slack-token" \
  --from-literal=discord-token="your-discord-token"
```

```bash
# 3. Deploy application
kubectl apply -k .
```

## Verify Deployment

```bash
# Check Pod status
kubectl get pods -n alert-webhooks

# Check service status
kubectl get svc -n alert-webhooks

# Check deployment status
kubectl get deployment -n alert-webhooks

# View Pod logs
kubectl logs -n alert-webhooks -l app=alert-webhooks

# View detailed information
kubectl describe deployment prod-alert-webhooks -n alert-webhooks

# Check Ingress status
kubectl get ingress -n alert-webhooks

# View Ingress details
kubectl describe ingress prod-alert-webhooks-internal -n alert-webhooks
```

## Update Deployment

```bash
# Update image version (modify newTag in kustomization.yaml)
# Then reapply
kubectl apply -k .

# Or set image directly
kubectl set image deployment/prod-alert-webhooks \
  alert-webhooks=docker_image_URL/alert-webhooks:{version} \
  -n alert-webhooks

# Rolling restart deployment
kubectl rollout restart deployment/prod-alert-webhooks -n alert-webhooks

# Check rollout status
kubectl rollout status deployment/prod-alert-webhooks -n alert-webhooks
```

## Troubleshooting

```bash
# Check Pod events
kubectl get events -n alert-webhooks --sort-by='.lastTimestamp'

# Debug Pod issues
kubectl describe pod -n alert-webhooks -l app=alert-webhooks

# Check ConfigMap
kubectl get configmap prod-alert-webhooks-config -n alert-webhooks -o yaml

# Check Secret
kubectl get secret alert-webhooks-secrets -n alert-webhooks -o yaml

# Port forward for local testing
kubectl port-forward -n alert-webhooks svc/prod-alert-webhooks-service 8080:80
```

## Configuration Management

### SSL Certificate and Domain

The SSL certificate ARN and domain name are managed through Kustomize patches in `kustomization.yaml`:

```yaml
patches:
  - target:
      kind: Ingress
      name: alert-webhooks-internal
    patch: |-
      - op: replace
        path: /metadata/annotations/alb.ingress.kubernetes.io~1certificate-arn
        value: arn:aws:acm:ap-northeast-1:{aws_account}:certificate/hash
      - op: replace
        path: /spec/rules/0/host
        value: alert-webhooks.domain.com
```

### Application Configuration

The application configuration is managed through ConfigMap generation from `configs/config.production.yaml`.

## Cleanup Resources

```bash
# Delete all resources
kubectl delete -k .

# Or delete entire namespace
kubectl delete namespace alert-webhooks
```

## Directory Structure

```
kustomize/
├── README.md
├── kustomization.yaml          # Main Kustomize configuration
├── base/                       # Base resources
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
└── configs/                    # Configuration files
    └── config.production.yaml
```
