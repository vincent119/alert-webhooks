# Kubernetes Environment Variables Configuration Guide

This guide explains how to use environment variables in Kubernetes to override sensitive configurations in `config.yaml`.

## Supported Environment Variables

The system reads configuration in the following priority order:

1. **Environment Variables** (highest priority)
2. `config.yaml` file (fallback)

## Authentication Systems

The system uses two independent authentication systems:

1. **Metrics Authentication** (`metric.user/password`):

   - Used for `/api/v1/metrics` endpoint (Prometheus metrics)
   - Only affects access to monitoring metrics

2. **Webhooks Authentication** (`webhooks.base_auth_user/password`):
   - Used for all Telegram/Slack API endpoints
   - Controls access to alert notification APIs
   - Authentication is only enabled when `webhooks.enable: true`

### Supported Environment Variables List

| Environment Variable | Configuration Path            | Description                                              |
| -------------------- | ----------------------------- | -------------------------------------------------------- |
| `METRIC_USER`        | `metric.user`                 | Username for Prometheus metrics endpoint authentication  |
| `METRIC_PASSWORD`    | `metric.password`             | Password for Prometheus metrics endpoint authentication  |
| `WEBHOOKS_USER`      | `webhooks.base_auth_user`     | Username for Telegram/Slack API endpoints authentication |
| `WEBHOOKS_PASSWORD`  | `webhooks.base_auth_password` | Password for Telegram/Slack API endpoints authentication |
| `TELEGRAM_TOKEN`     | `telegram.token`              | Telegram Bot API Token                                   |
| `SLACK_TOKEN`        | `slack.token`                 | Slack Bot API Token                                      |

## Kubernetes Deployment Example

### 1. Create Secret

```bash
# Method 1: Create secret using kubectl
kubectl create secret generic alert-webhooks-secrets \
  --from-literal=metric-user="your-metric-user" \
  --from-literal=metric-password="your-metric-password" \
  --from-literal=webhooks-user="your-webhook-user" \
  --from-literal=webhooks-password="your-webhook-password" \
  --from-literal=telegram-token="your-telegram-bot-token" \
  --from-literal=slack-token="your-slack-bot-token" \
  --namespace monitoring

# Method 2: Manual encoding and apply YAML
echo -n "your-value" | base64
```

### 2. Reference Secret in Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alert-webhooks
spec:
  template:
    spec:
      containers:
        - name: alert-webhooks
          env:
            - name: METRIC_USER
              valueFrom:
                secretKeyRef:
                  name: alert-webhooks-secrets
                  key: metric-user
            - name: METRIC_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: alert-webhooks-secrets
                  key: metric-password
          # ... other environment variables
```

## Configuration File Example

### config.production.yaml

```yaml
metric:
  # Default values used when environment variables are not set
  user: "default-user"
  password: "default-password"

webhooks:
  # enable field defaults to false, set to true explicitly to enable
  enable: true
  # Default values used when environment variables are not set
  base_auth_user: "default-user"
  base_auth_password: "default-password"

telegram:
  # enable field defaults to false, set to true explicitly to enable
  enable: true
  # Default value used when TELEGRAM_TOKEN environment variable is not set
  token: "default-token"

slack:
  # enable field defaults to false, set to true explicitly to enable
  enable: true
  # Default value used when SLACK_TOKEN environment variable is not set
  token: "default-token"
```

## Service Enable Logic

**Important**: All services (webhooks, telegram, slack) have `enable` field default to `false`.

- **If `enable` field is missing from config file**: Service will be disabled
- **If `enable` field is empty in config file**: Service will be disabled
- **If `enable: false` in config file**: Service will be disabled
- **If `enable: true` in config file**: Service will be enabled

This design ensures security - only explicitly configured services will run.

## Security Best Practices

1. **Use Kubernetes Secrets**: All sensitive information should be stored in Kubernetes Secrets
2. **Restrict Secret Access**: Use RBAC to control who can access these Secrets
3. **Encrypt Data at Rest**: Ensure Secrets data in etcd is encrypted
4. **Regular Token Rotation**: Regularly update Bot Tokens and passwords

## Verify Configuration

After deployment, you can verify that environment variables are loaded correctly:

1. Check application logs:

```bash
kubectl logs deployment/alert-webhooks -n monitoring
```

2. Look for these log messages:

```
Override metric user from env var: your-user
Override metric password from env var: [REDACTED]
Override telegram token from env var: [REDACTED]
Override slack token from env var: [REDACTED]
Override webhooks user from env var: your-user
Override webhooks password from env var: [REDACTED]
```

## Troubleshooting

### Environment Variables Not Taking Effect

1. Check if Secret exists: `kubectl get secrets -n monitoring`
2. Check Secret contents: `kubectl describe secret alert-webhooks-secrets -n monitoring`
3. Check Pod environment variables: `kubectl exec -it <pod-name> -- env | grep -E "(METRIC|WEBHOOK|TELEGRAM|SLACK)"`

### Pod Cannot Start

1. Check if Secret references are correct
2. Check if ConfigMap exists
3. Check Pod events: `kubectl describe pod <pod-name> -n monitoring`

## Complete Example

See `kubernetes/deployment-example.yaml` file for a complete Kubernetes deployment example.

## Related Documentation

- [Kubernetes Basic Configuration](./kubernetes-basic.md)
- [Telegram Setup Guide](./telegram_setup.md)
- [Slack Setup Guide](./slack_setup.md)
