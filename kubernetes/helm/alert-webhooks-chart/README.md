# alert-webhooks-chart

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0](https://img.shields.io/badge/AppVersion-v1.0-informational?style=flat-square)

A Helm chart for Alert Webhooks - Monitoring notification service

**Homepage:** <https://github.com/vincent119/alert-webhooks>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| DevOps Team | <vincent119@gmail.com> |  |

## Source Code

* <https://github.com/vincent119/alert-webhooks>

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
| app.version | string | `"1.0.0"` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `5` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` |  |
| config.app.mode | string | `"production"` |  |
| config.app.port | string | `"9999"` |  |
| config.app.version | string | `"1.0.0"` |  |
| config.discord.avatar_url | string | `""` |  |
| config.discord.channels.chat_ids0 | string | `"ID"` |  |
| config.discord.channels.chat_ids1 | string | `"ID"` |  |
| config.discord.channels.chat_ids2 | string | `"ID"` |  |
| config.discord.channels.chat_ids3 | string | `"ID"` |  |
| config.discord.channels.chat_ids4 | string | `"ID"` |  |
| config.discord.channels.chat_ids5 | string | `"ID"` |  |
| config.discord.enable | bool | `true` |  |
| config.discord.guild_id | string | `"ID"` |  |
| config.discord.template_language | string | `"tw"` |  |
| config.discord.template_mode | string | `"full"` |  |
| config.discord.username | string | `"AlertBot"` |  |
| config.log.format | string | `"json"` |  |
| config.log.level | string | `"info"` |  |
| config.slack.channel | string | `"#alerts"` |  |
| config.slack.channels.chat_ids0 | string | `"#alert0"` |  |
| config.slack.channels.chat_ids1 | string | `"#alert1"` |  |
| config.slack.channels.chat_ids2 | string | `"#alert2"` |  |
| config.slack.channels.chat_ids3 | string | `"#alert3"` |  |
| config.slack.channels.chat_ids4 | string | `"#alert4"` |  |
| config.slack.channels.chat_ids5 | string | `"#alert5"` |  |
| config.slack.enable | bool | `true` |  |
| config.slack.icon_emoji | string | `":warning:"` |  |
| config.slack.link_names | bool | `true` |  |
| config.slack.template_language | string | `"tw"` |  |
| config.slack.template_mode | string | `"full"` |  |
| config.slack.unfurl_links | bool | `false` |  |
| config.slack.unfurl_media | bool | `false` |  |
| config.slack.username | string | `"AlertBot"` |  |
| config.telegram.chat_ids[0] | string | `"ID"` |  |
| config.telegram.chat_ids[1] | string | `"ID"` |  |
| config.telegram.chat_ids[2] | string | `"ID"` |  |
| config.telegram.chat_ids[3] | string | `"ID"` |  |
| config.telegram.chat_ids[4] | string | `"ID"` |  |
| config.telegram.chat_ids[5] | string | `"ID"` |  |
| config.telegram.chat_ids[6] | string | `"ID"` |  |
| config.telegram.enable | bool | `true` |  |
| config.telegram.template_language | string | `"tw"` |  |
| config.telegram.template_mode | string | `"full"` |  |
| config.webhooks.enable | bool | `true` |  |
| env[0].name | string | `"SWAGGER_HOST"` |  |
| env[0].value | string | `"alert-webhooks.domain.com"` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"vincent119/alert-webhooks"` |  |
| image.tag | string | `"v1.0.0"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations."alb.ingress.kubernetes.io/backend-protocol" | string | `"HTTP"` |  |
| ingress.annotations."alb.ingress.kubernetes.io/certificate-arn" | string | `"arn:aws:acm:ap-northeast-1:{aws_account}:certificate/hash"` |  |
| ingress.annotations."alb.ingress.kubernetes.io/group.name" | string | `"infra"` |  |
| ingress.annotations."alb.ingress.kubernetes.io/listen-ports" | string | `"[{\"HTTP\": 80}, {\"HTTPS\": 443}]"` |  |
| ingress.annotations."alb.ingress.kubernetes.io/load-balancer-attributes" | string | `"deletion_protection.enabled=true,idle_timeout.timeout_seconds=60"` |  |
| ingress.annotations."alb.ingress.kubernetes.io/load-balancer-name" | string | `"infra"` |  |
| ingress.annotations."alb.ingress.kubernetes.io/target-type" | string | `"ip"` |  |
| ingress.className | string | `"alb"` |  |
| ingress.enabled | bool | `true` |  |
| ingress.hosts[0].host | string | `"alert-webhooks.domain.com"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
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
| secrets.discord.token | string | `"your-discord-token"` |  |
| secrets.metric.password | string | `"admin"` |  |
| secrets.metric.user | string | `"admin"` |  |
| secrets.slack.token | string | `"your-slack-token"` |  |
| secrets.telegram.token | string | `"your-telegram-token"` |  |
| secrets.webhooks.password | string | `"admin"` |  |
| secrets.webhooks.user | string | `"admin"` |  |
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
