{{/*
Expand the name of the chart.
*/}}
{{- define "pr-preview.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name using the release name.
*/}}
{{- define "pr-preview.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to all resources.
*/}}
{{- define "pr-preview.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
trainyard/pr-number: {{ .Values.global.prNumber | quote }}
{{- end }}

{{/*
Selector labels for a specific service.
Usage: include "pr-preview.selectorLabels" (dict "name" "web" "Release" .Release)
*/}}
{{- define "pr-preview.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Resolve the subdomain for a service.
Replaces {number} with the actual PR number.
Usage: include "pr-preview.subdomain" (dict "subdomain" "pr-{number}" "prNumber" "42")
*/}}
{{- define "pr-preview.subdomain" -}}
{{- .subdomain | replace "{number}" (.prNumber | toString) }}
{{- end }}

{{/*
Full hostname for a public service.
Usage: include "pr-preview.hostname" (dict "subdomain" "pr-{number}" "prNumber" "42" "domain" "preview.example.com")
*/}}
{{- define "pr-preview.hostname" -}}
{{- $subdomain := include "pr-preview.subdomain" (dict "subdomain" .subdomain "prNumber" .prNumber) }}
{{- printf "%s.%s" $subdomain .domain }}
{{- end }}
