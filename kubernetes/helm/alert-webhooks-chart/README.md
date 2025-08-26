# Alert Webhooks Helm Chart

A Helm chart for deploying Alert Webhooks - a monitoring notification service that supports Telegram, Slack, and Discord integrations.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- AWS Load Balancer Controller (for ALB Ingress)

## Installation

### 1. Add your secrets

Create a `values-production.yaml` file with your actual secrets:

```yaml
secrets:
  metric:
    user: "your-metric-user"
    password: "your-metric-password"
  webhooks:
    user: "your-webhooks-user"
    password: "your-webhooks-password"
  telegram:
    token: "your-telegram-bot-token"
  slack:
    token: "your-slack-bot-token"
  discord:
    token: "your-discord-bot-token"

ingress:
  annotations:
    alb.ingress.kubernetes.io/certificate-arn: "arn:aws:acm:region:account:certificate/your-cert-id"
  hosts:
    - host: your-domain.com
      paths:
        - path: /
          pathType: Prefix
```

### 2. Install the chart

```bash
# Create namespace
kubectl create namespace alert-webhooks

# Install with custom values
helm install alert-webhooks . \
  --namespace alert-webhooks \
  --values values-production.yaml

# Or install with inline values
helm install alert-webhooks . \
  --namespace alert-webhooks \
  --set secrets.telegram.token="your-token" \
  --set secrets.slack.token="your-token" \
  --set secrets.discord.token="your-token"
```

### 3. Verify installation

```bash
# Check deployment status
helm status alert-webhooks -n alert-webhooks

# Check pods
kubectl get pods -n alert-webhooks

# Check services
kubectl get svc -n alert-webhooks

# Check ingress
kubectl get ingress -n alert-webhooks
```

## Configuration

### Application Configuration

The chart supports comprehensive configuration through `values.yaml`:

- **Replicas**: Configure replica count and autoscaling
- **Image**: Set container image repository and tag
- **Resources**: Define CPU and memory limits/requests
- **Health Checks**: Configure liveness and readiness probes

### Service Configuration

Configure each notification service:

```yaml
config:
  telegram:
    enable: true
    chat_ids: ["-1001234567890", "-1001234567891"]
    template_mode: "full"
    template_language: "tw"

  slack:
    enable: true
    channels:
      chat_ids0: "#alerts-critical"
      chat_ids1: "#alerts-warning"

  discord:
    enable: true
    guild_id: "your-guild-id"
    channels:
      chat_ids0: "channel-id-1"
      chat_ids1: "channel-id-2"
```

### Ingress Configuration

Configure AWS ALB Ingress:

```yaml
ingress:
  enabled: true
  className: "alb"
  annotations:
    alb.ingress.kubernetes.io/load-balancer-name: infra
    alb.ingress.kubernetes.io/certificate-arn: "your-cert-arn"
    alb.ingress.kubernetes.io/target-type: "ip"
  hosts:
    - host: alert-webhooks.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
```

## Upgrading

```bash
# Update chart with new values
helm upgrade alert-webhooks . \
  --namespace alert-webhooks \
  --values values-production.yaml

# Update image tag
helm upgrade alert-webhooks . \
  --namespace alert-webhooks \
  --set image.tag=v1.1
```

## Uninstallation

```bash
# Uninstall the chart
helm uninstall alert-webhooks -n alert-webhooks

# Delete namespace (optional)
kubectl delete namespace alert-webhooks
```

## Troubleshooting

### Check Pod Logs

```bash
kubectl logs -n alert-webhooks -l app.kubernetes.io/name=alert-webhooks-chart
```

### Check ConfigMap

```bash
kubectl get configmap -n alert-webhooks -o yaml
```

### Check Secrets

```bash
kubectl get secret -n alert-webhooks
kubectl describe secret alert-webhooks-secrets -n alert-webhooks
```

### Debug Ingress

```bash
kubectl describe ingress -n alert-webhooks
kubectl get events -n alert-webhooks --sort-by='.lastTimestamp'
```

## Values

| Key                       | Type   | Default                             | Description        |
| ------------------------- | ------ | ----------------------------------- | ------------------ |
| `replicaCount`            | int    | `1`                                 | Number of replicas |
| `image.repository`        | string | `"docker-replotory/alert-webhooks"` | Image repository   |
| `image.tag`               | string | `"v1.0"`                            | Image tag          |
| `service.type`            | string | `"ClusterIP"`                       | Service type       |
| `service.port`            | int    | `80`                                | Service port       |
| `service.targetPort`      | int    | `9999`                              | Container port     |
| `ingress.enabled`         | bool   | `true`                              | Enable ingress     |
| `ingress.className`       | string | `"alb"`                             | Ingress class name |
| `resources.limits.cpu`    | string | `"500m"`                            | CPU limit          |
| `resources.limits.memory` | string | `"512Mi"`                           | Memory limit       |
| `autoscaling.enabled`     | bool   | `false`                             | Enable autoscaling |

## Support

For issues and questions:

- Check the [troubleshooting section](#troubleshooting)
- Review pod logs and events
- Verify configuration values
