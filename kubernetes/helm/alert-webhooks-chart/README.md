# alert-webhooks-chart

![Version: 1.0.2](https://img.shields.io/badge/Version-1.0.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.0.10](https://img.shields.io/badge/AppVersion-v2.0.10-informational?style=flat-square)

A Helm chart for Alert Webhooks - Monitoring notification service

**Homepage:** <https://github.com/vincent119/alert-webhooks>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| vincent119 | <vincent119@gmail.com> |  |

## Source Code

* <https://github.com/vincent119/alert-webhooks>

## Prerequisites

- A reachable Kubernetes cluster and a kubeconfig with access to the target namespace.
- Helm and kubectl. On macOS, install them with Homebrew:

```bash
brew install helm kubectl
helm version
kubectl version --client
kubectl config current-context
```

For other operating systems, see the [Helm installation guide](https://helm.sh/docs/intro/install/) and [kubectl installation guide](https://kubernetes.io/docs/tasks/tools/). Before deployment, confirm that `kubectl config current-context` points to the intended cluster.

## Security configuration

The Chart does not provide default webhook, metric, or notification service credentials. Supply `secrets.metric.user`, `secrets.metric.password`, `secrets.webhooks.user`, and `secrets.webhooks.password` through a private values file.

When `config.telegram.enable`, `config.slack.enable`, or `config.discord.enable` is `true`, provide the corresponding `secrets.<service>.token`. Trace authentication remains optional.

Do not commit real credentials to the repository or pass secrets through `--set`, which can leave them in shell history or CI logs. Store the private values file outside the repository and restrict its file permissions.

The following minimum example disables all notification services. Add tokens to the private values file only when enabling a service:

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
- A `chart-v<Chart version>` tag publishes only the matching immutable Chart package.
- This release uses Chart version `1.0.2`, application version `v2.0.10`, and container image tag `v2.0.10` by default.
- Published Chart versions are not overwritten, and the repository does not provide a mutable `latest` Chart.

## Install from the public repository

Add the public Helm repository and select an explicit Chart version:

```bash
helm repo add alert-webhooks https://vincent119.github.io/alert-webhooks
helm repo update
helm show chart alert-webhooks/alert-webhooks-chart --version 1.0.2
```

Set the path to a private values file outside the repository:

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
```

Render the Chart without modifying the cluster:

```bash
helm template alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.2 \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

Review the rendered output for placeholders, unexpected image references, and secrets before continuing.

Install or upgrade after the rendered output has been reviewed:

```bash
helm upgrade --install alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.2 \
  --namespace alert-webhooks \
  --create-namespace \
  --values "${PRIVATE_VALUES_FILE}" \
  --atomic \
  --wait \
  --timeout 5m
```

## Validate from source

Run the following commands from the repository root:

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
CHART_DIR=kubernetes/helm/alert-webhooks-chart

helm lint "${CHART_DIR}" --values "${PRIVATE_VALUES_FILE}"
helm template alert-webhooks "${CHART_DIR}" \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

## Verify the deployment

```bash
helm status alert-webhooks --namespace alert-webhooks
kubectl get pods,service,ingress --namespace alert-webhooks
kubectl rollout status deployment \
  --namespace alert-webhooks \
  --selector app.kubernetes.io/instance=alert-webhooks \
  --timeout 5m
```

Do not continue with notification testing until the rollout completes and the Pods are ready.

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

`helm uninstall` does not delete the namespace automatically. Before deleting the namespace, confirm that it contains no unrelated workloads or retained resources.

## Maintain this README

Maintain the hand-written content in `README.md.gotmpl`. Generate the values table with the repository-pinned `helm-docs` version:

```bash
make helm-docs
```

Do not edit the generated values table manually. The [Traditional Chinese guide](./README.zh-TW.md) provides the same installation and lifecycle overview without duplicating the generated values reference.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| alertTemplateConfig.template_config.fallback_order[0] | string | `"eng"` |  |
| alertTemplateConfig.template_config.fallback_order[1] | string | `"tw"` |  |
| alertTemplateConfig.template_config.fallback_order[2] | string | `"zh"` |  |
| alertTemplateConfig.template_config.fallback_order[3] | string | `"en"` |  |
| alertTemplateConfig.template_config.format_options.compact_mode.description | string | `"Compact mode (simplified display)"` |  |
| alertTemplateConfig.template_config.format_options.compact_mode.enabled | bool | `false` |  |
| alertTemplateConfig.template_config.format_options.max_summary_length.description | string | `"Maximum summary length"` |  |
| alertTemplateConfig.template_config.format_options.max_summary_length.value | int | `200` |  |
| alertTemplateConfig.template_config.format_options.show_emoji.description | string | `"Whether to show emojis"` |  |
| alertTemplateConfig.template_config.format_options.show_emoji.enabled | bool | `true` |  |
| alertTemplateConfig.template_config.format_options.show_external_url.description | string | `"Whether to show external URLs"` |  |
| alertTemplateConfig.template_config.format_options.show_external_url.enabled | bool | `true` |  |
| alertTemplateConfig.template_config.format_options.show_generator_url.description | string | `"Whether to show generator URLs"` |  |
| alertTemplateConfig.template_config.format_options.show_generator_url.enabled | bool | `true` |  |
| alertTemplateConfig.template_config.format_options.show_links.description | string | `"Whether to show hyperlinks"` |  |
| alertTemplateConfig.template_config.format_options.show_links.enabled | bool | `true` |  |
| alertTemplateConfig.template_config.format_options.show_timestamps.description | string | `"Whether to show timestamps"` |  |
| alertTemplateConfig.template_config.format_options.show_timestamps.enabled | bool | `true` |  |
| alertTemplateConfig.template_config.naming_convention.prefix | string | `"alert_template_"` |  |
| alertTemplateConfig.template_config.naming_convention.priority_order[0] | string | `".tmpl"` |  |
| alertTemplateConfig.template_config.naming_convention.priority_order[1] | string | `".j2"` |  |
| alertTemplateConfig.template_config.naming_convention.supported_extensions[0] | string | `".tmpl"` |  |
| alertTemplateConfig.template_config.naming_convention.supported_extensions[1] | string | `".j2"` |  |
| alertTemplateConfig.template_config.supported_languages[0].code | string | `"eng"` |  |
| alertTemplateConfig.template_config.supported_languages[0].description | string | `"English template for alert notifications"` |  |
| alertTemplateConfig.template_config.supported_languages[0].fallback | bool | `true` |  |
| alertTemplateConfig.template_config.supported_languages[0].name | string | `"English"` |  |
| alertTemplateConfig.template_config.supported_languages[1].code | string | `"tw"` |  |
| alertTemplateConfig.template_config.supported_languages[1].description | string | `"Traditional Chinese template for alert notifications"` |  |
| alertTemplateConfig.template_config.supported_languages[1].fallback | bool | `false` |  |
| alertTemplateConfig.template_config.supported_languages[1].name | string | `"繁體中文"` |  |
| alertTemplateConfig.template_config.supported_languages[2].code | string | `"zh"` |  |
| alertTemplateConfig.template_config.supported_languages[2].description | string | `"Simplified Chinese template for alert notifications"` |  |
| alertTemplateConfig.template_config.supported_languages[2].fallback | bool | `false` |  |
| alertTemplateConfig.template_config.supported_languages[2].name | string | `"简体中文"` |  |
| alertTemplateConfig.template_config.supported_languages[3].code | string | `"ko"` |  |
| alertTemplateConfig.template_config.supported_languages[3].description | string | `"Korean template for alert notifications"` |  |
| alertTemplateConfig.template_config.supported_languages[3].fallback | bool | `false` |  |
| alertTemplateConfig.template_config.supported_languages[3].name | string | `"한국어"` |  |
| alertTemplateConfig.template_config.supported_languages[4].code | string | `"ja"` |  |
| alertTemplateConfig.template_config.supported_languages[4].description | string | `"Japanese template for alert notifications"` |  |
| alertTemplateConfig.template_config.supported_languages[4].fallback | bool | `false` |  |
| alertTemplateConfig.template_config.supported_languages[4].name | string | `"日本語"` |  |
| alertTemplateConfig.template_config.version | string | `"1.0.0"` |  |
| alertTemplateConfigMinimal.template_config.fallback_order[0] | string | `"eng"` |  |
| alertTemplateConfigMinimal.template_config.fallback_order[1] | string | `"tw"` |  |
| alertTemplateConfigMinimal.template_config.fallback_order[2] | string | `"zh"` |  |
| alertTemplateConfigMinimal.template_config.format_options.compact_mode.description | string | `"Compact mode (simplified display)"` |  |
| alertTemplateConfigMinimal.template_config.format_options.compact_mode.enabled | bool | `true` |  |
| alertTemplateConfigMinimal.template_config.format_options.max_summary_length.description | string | `"Maximum summary length"` |  |
| alertTemplateConfigMinimal.template_config.format_options.max_summary_length.value | int | `100` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_emoji.description | string | `"Whether to show emojis"` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_emoji.enabled | bool | `true` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_external_url.description | string | `"Whether to show external URLs"` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_external_url.enabled | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_generator_url.description | string | `"Whether to show generator URLs"` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_generator_url.enabled | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_links.description | string | `"Whether to show hyperlinks"` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_links.enabled | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_timestamps.description | string | `"Whether to show timestamps"` |  |
| alertTemplateConfigMinimal.template_config.format_options.show_timestamps.enabled | bool | `true` |  |
| alertTemplateConfigMinimal.template_config.naming_convention.prefix | string | `"alert_template_"` |  |
| alertTemplateConfigMinimal.template_config.naming_convention.priority_order[0] | string | `".tmpl"` |  |
| alertTemplateConfigMinimal.template_config.naming_convention.priority_order[1] | string | `".j2"` |  |
| alertTemplateConfigMinimal.template_config.naming_convention.supported_extensions[0] | string | `".tmpl"` |  |
| alertTemplateConfigMinimal.template_config.naming_convention.supported_extensions[1] | string | `".j2"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[0].code | string | `"eng"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[0].description | string | `"English template for alert notifications"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[0].fallback | bool | `true` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[0].name | string | `"English"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[1].code | string | `"tw"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[1].description | string | `"Traditional Chinese template for alert notifications"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[1].fallback | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[1].name | string | `"繁體中文"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[2].code | string | `"zh"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[2].description | string | `"Simplified Chinese template for alert notifications"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[2].fallback | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[2].name | string | `"简体中文"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[3].code | string | `"ko"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[3].description | string | `"Korean template for alert notifications"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[3].fallback | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[3].name | string | `"한국어"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[4].code | string | `"ja"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[4].description | string | `"Japanese template for alert notifications"` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[4].fallback | bool | `false` |  |
| alertTemplateConfigMinimal.template_config.supported_languages[4].name | string | `"日本語"` |  |
| alertTemplateConfigMinimal.template_config.version | string | `"1.0.0"` |  |
| app.environment | string | `"production"` |  |
| app.name | string | `"alert-webhooks"` |  |
| app.version | string | `"v2.0.10"` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `5` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` |  |
| config.app.mode | string | `"production"` |  |
| config.app.port | string | `"9999"` |  |
| config.app.version | string | `"v2.0.10"` |  |
| config.discord.avatar_url | string | `""` |  |
| config.discord.channels | object | `{}` |  |
| config.discord.enable | bool | `false` |  |
| config.discord.guild_id | string | `""` |  |
| config.discord.template_language | string | `"tw"` |  |
| config.discord.template_mode | string | `"full"` |  |
| config.discord.username | string | `"AlertBot"` |  |
| config.log.format | string | `"json"` |  |
| config.log.level | string | `"info"` |  |
| config.slack.channel | string | `""` |  |
| config.slack.channels | object | `{}` |  |
| config.slack.enable | bool | `false` |  |
| config.slack.icon_emoji | string | `":warning:"` |  |
| config.slack.link_names | bool | `true` |  |
| config.slack.template_language | string | `"tw"` |  |
| config.slack.template_mode | string | `"full"` |  |
| config.slack.unfurl_links | bool | `false` |  |
| config.slack.unfurl_media | bool | `false` |  |
| config.slack.username | string | `"AlertBot"` |  |
| config.telegram.chat_ids | list | `[]` |  |
| config.telegram.enable | bool | `false` |  |
| config.telegram.template_language | string | `"tw"` |  |
| config.telegram.template_mode | string | `"full"` |  |
| config.trace.authPasswd | string | `""` |  |
| config.trace.authUser | string | `""` |  |
| config.trace.enable | bool | `false` |  |
| config.trace.insecure | bool | `false` |  |
| config.trace.port | string | `""` |  |
| config.trace.sampleRate | float | `1` |  |
| config.trace.tlsSkipVerify | bool | `false` |  |
| config.trace.url | string | `""` |  |
| config.trace.urlPath | string | `""` |  |
| config.webhooks.enable | bool | `true` |  |
| env[0].name | string | `"SWAGGER_HOST"` |  |
| env[0].value | string | `"localhost:9999"` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"docker.io/vincent119/alert-webhooks"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | list | `[]` |  |
| ingress.tls | list | `[]` |  |
| initContainers | list | `[]` |  |
| livenessProbe.failureThreshold | int | `3` |  |
| livenessProbe.httpGet.path | string | `"/healthy"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| livenessProbe.initialDelaySeconds | int | `60` |  |
| livenessProbe.periodSeconds | int | `30` |  |
| livenessProbe.timeoutSeconds | int | `10` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| readinessProbe.failureThreshold | int | `3` |  |
| readinessProbe.httpGet.path | string | `"/healthy"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.initialDelaySeconds | int | `10` |  |
| readinessProbe.periodSeconds | int | `10` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | string | `"500m"` |  |
| resources.limits.memory | string | `"512Mi"` |  |
| resources.requests.cpu | string | `"200m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| secrets.discord.token | string | `""` |  |
| secrets.metric.password | string | `""` |  |
| secrets.metric.user | string | `""` |  |
| secrets.slack.token | string | `""` |  |
| secrets.telegram.token | string | `""` |  |
| secrets.trace.authPasswd | string | `""` |  |
| secrets.trace.authUser | string | `""` |  |
| secrets.webhooks.password | string | `""` |  |
| secrets.webhooks.user | string | `""` |  |
| securityContext | object | `{}` |  |
| service.port | int | `80` |  |
| service.targetPort | int | `9999` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |
| volumeMounts | list | `[]` |  |
| volumes | list | `[]` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
