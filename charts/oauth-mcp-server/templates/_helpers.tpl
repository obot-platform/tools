{{/*
Return the chart name and version.
*/}}
{{- define "mcpServer.chart" -}}
{{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "mcpServer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a fullname using the release name and the chart name.
*/}}
{{- define "mcpServer.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "mcpServer.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create labels for the resources.
*/}}
{{- define "mcpServer.labels" -}}
helm.sh/chart: {{ include "mcpServer.chart" . }}
{{ include "mcpServer.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create selector labels for the resources.
*/}}
{{- define "mcpServer.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mcpServer.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "mcpServer.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "mcpServer.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Set name of secret to use for mcp server
*/}}
{{- define "mcpServer.config.secretName" -}}
{{- if .Values.mcpServer.config.existingSecret -}}
{{- .Values.mcpServer.config.existingSecret -}}
{{- else -}}
{{ .Release.Name }}-mcp-server-config
{{- end -}}
{{- end -}}

{{/*
Set name of secret to use for proxy
*/}}
{{- define "proxy.config.secretName" -}}
{{- if .Values.proxy.config.existingSecret -}}
{{- .Values.proxy.config.existingSecret -}}
{{- else -}}
{{ .Release.Name }}-proxy-config
{{- end -}}
{{- end -}}