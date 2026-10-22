# alert-webhooks-chart

![Version: 1.0.1](https://img.shields.io/badge/Version-1.0.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.0.8](https://img.shields.io/badge/AppVersion-v2.0.8-informational?style=flat-square)

A Helm chart for Alert Webhooks - Monitoring notification service

**Homepage:** <https://github.com/vincent119/alert-webhooks>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| vincent119 | <vincent119@gmail.com> |  |

## Source Code

* <https://github.com/vincent119/alert-webhooks>

## 前置條件

- 可連線的 Kubernetes cluster，以及具備目標 namespace 權限的 kubeconfig。
- Helm 與 kubectl。macOS 可使用 Homebrew 安裝：

```bash
brew install helm kubectl
helm version
kubectl version --client
kubectl config current-context
```

其他作業系統請參考 [Helm 安裝文件](https://helm.sh/docs/intro/install/) 與 [kubectl 安裝文件](https://kubernetes.io/docs/tasks/tools/)。執行部署前，必須確認 `kubectl config current-context` 指向正確 cluster。

## 安全設定

Chart 不提供 webhook、metric 或通知服務 Token 的預設值。安裝時必須以私有 values 檔提供 `secrets.metric.user`、`secrets.metric.password`、`secrets.webhooks.user` 與 `secrets.webhooks.password`。

當 `config.telegram.enable`、`config.slack.enable` 或 `config.discord.enable` 為 `true` 時，必須分別提供對應的 `secrets.<service>.token`。Trace 認證維持選填。

請勿將實際憑證提交至 repository，也不要透過 `--set` 把 Secret 留在 shell history 或 CI log。私有 values 檔應放在 repository 外，並限制檔案權限。

以下最小範例停用通知服務；需要啟用時再於私有 values 檔加入對應 Token：

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

## 版本契約

- Chart version 與 application version 分開管理。
- `chart-v<Chart version>` tag 只發布對應的不可變 Chart package。
- 本版本的 Chart version 為 `1.0.1`，`appVersion` 與預設 container image tag 為 `v2.0.8`。
- 公開 repository 不提供 `latest` Chart，也不覆寫已發布的版本。

## 從公開 repository 安裝

先加入公開 Helm repository，並固定使用所需 Chart version：

```bash
helm repo add alert-webhooks https://vincent119.github.io/alert-webhooks
helm repo update
helm show chart alert-webhooks/alert-webhooks-chart --version 1.0.1
```

指定 repository 外的私有 values 路徑：

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
```

先執行 render 驗證，不連線修改 cluster：

```bash
helm template alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.1 \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

確認 render 結果後安裝或升級：

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

## 從原始碼驗證

以下命令預設從 repository root 執行：

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
CHART_DIR=kubernetes/helm/alert-webhooks-chart

helm lint "${CHART_DIR}" --values "${PRIVATE_VALUES_FILE}"
helm template alert-webhooks "${CHART_DIR}" \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

## 驗證部署

```bash
helm status alert-webhooks --namespace alert-webhooks
kubectl get pods,service,ingress --namespace alert-webhooks
kubectl rollout status deployment \
  --namespace alert-webhooks \
  --selector app.kubernetes.io/instance=alert-webhooks \
  --timeout 5m
```

## 回滾與移除

先查看 release revision，再指定實際 revision 回滾：

```bash
helm history alert-webhooks --namespace alert-webhooks
REVISION_NUMBER=1
helm rollback alert-webhooks "${REVISION_NUMBER}" \
  --namespace alert-webhooks \
  --wait \
  --timeout 5m
```

移除 release：

```bash
helm uninstall alert-webhooks --namespace alert-webhooks
```

`helm uninstall` 不會自動刪除 namespace。刪除 namespace 前，先確認其中沒有其他工作負載。

## 維護 README

人工內容維護於 `README.md.gotmpl`，values 表格由固定版本的 `helm-docs` 產生：

```bash
make helm-docs
```

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
| app.version | string | `"v2.0.8"` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `5` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` |  |
| config.app.mode | string | `"production"` |  |
| config.app.port | string | `"9999"` |  |
| config.app.version | string | `"v2.0.8"` |  |
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
