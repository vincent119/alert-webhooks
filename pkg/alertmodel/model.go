package alertmodel

import (
	"alert-webhooks/pkg/template"
)

// BuildTemplateData 將 AlertManager 直傳格式（map 結構）轉換為模板引擎可用的 TemplateData
// - status: 總體狀態（firing/resolved）
// - alerts: 警報列表，使用通用 map 結構（需包含 status/labels/annotations/startsAt/endsAt/generatorURL）
// - groupLabels/commonLabels/commonAnnotations: 群組與通用標籤/註解（map[string]interface{}）
// - externalURL: Alertmanager 外部 URL
// - formatOptions: 模板格式化選項（由各平台 handler 透過 templateEngine 決定）
func BuildTemplateData(
    status string,
    alerts []map[string]interface{},
    groupLabels map[string]interface{},
    commonLabels map[string]interface{},
    commonAnnotations map[string]interface{},
    externalURL string,
    formatOptions template.FormatOptions,
) template.TemplateData {
    // 統計 firing / resolved 數量
    firingCount := 0
    resolvedCount := 0
    for _, alert := range alerts {
        if s, ok := alert["status"].(string); ok {
            if s == "firing" {
                firingCount++
            } else if s == "resolved" {
                resolvedCount++
            }
        }
    }

    // 從 CommonLabels 讀取基本欄位，若沒有則從第一筆 alert 的 labels 回退
    var alertName, env, severity, namespace string
    if commonLabels != nil {
        if v, ok := commonLabels["alertname"].(string); ok {
            alertName = v
        }
        if v, ok := commonLabels["env"].(string); ok {
            env = v
        }
        if v, ok := commonLabels["severity"].(string); ok {
            severity = v
        }
        if v, ok := commonLabels["namespace"].(string); ok {
            namespace = v
        }
    }

    // 若欄位仍然為空，從第一筆 alert.labels 回退
    if len(alerts) > 0 {
        if labels, ok := alerts[0]["labels"].(map[string]interface{}); ok {
            if alertName == "" {
                if v, ok := labels["alertname"].(string); ok {
                    alertName = v
                }
            }
            if env == "" {
                if v, ok := labels["env"].(string); ok {
                    env = v
                }
            }
            if severity == "" {
                if v, ok := labels["severity"].(string); ok {
                    severity = v
                }
            }
            if namespace == "" {
                if v, ok := labels["namespace"].(string); ok {
                    namespace = v
                }
            }
        }
    }

    // 轉換 alerts 為模板引擎使用的結構
    var alertData []template.AlertData
    for _, a := range alerts {
        item := template.AlertData{
            Labels:      map[string]string{},
            Annotations: map[string]string{},
        }
        if v, ok := a["status"].(string); ok { item.Status = v }
        if v, ok := a["startsAt"].(string); ok { item.StartsAt = v }
        if v, ok := a["endsAt"].(string); ok { item.EndsAt = v }
        if v, ok := a["generatorURL"].(string); ok { item.GeneratorURL = v }

        if labels, ok := a["labels"].(map[string]interface{}); ok {
            for k, val := range labels {
                if s, ok := val.(string); ok {
                    item.Labels[k] = s
                }
            }
        }

        if ann, ok := a["annotations"].(map[string]interface{}); ok {
            for k, val := range ann {
                if s, ok := val.(string); ok {
                    item.Annotations[k] = s
                }
            }
        }

        alertData = append(alertData, item)
    }

    // 組裝 TemplateData，並帶入外部連結與格式選項
    data := template.TemplateData{
        Status:        status,
        AlertName:     alertName,
        Env:           env,
        Severity:      severity,
        Namespace:     namespace,
        TotalAlerts:   len(alerts),
        FiringCount:   firingCount,
        ResolvedCount: resolvedCount,
        Alerts:        alertData,
        ExternalURL:   externalURL,
        FormatOptions: formatOptions,
    }

    return data
}




