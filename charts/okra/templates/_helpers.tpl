{{/*
Expand the name of the chart.
*/}}
{{- define "okra.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "okra.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "okra.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "okra.labels" -}}
helm.sh/chart: {{ include "okra.chart" . }}
{{ include "okra.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "okra.selectorLabels" -}}
app.kubernetes.io/name: {{ include "okra.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "okra.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "okra.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "okra.leaderElectionRoleName" -}}
{{- include "okra.fullname" . }}-leader-election
{{- end }}

{{- define "okra.authProxyRoleName" -}}
{{- include "okra.fullname" . }}-proxy
{{- end }}

{{- define "okra.managerRoleName" -}}
{{- include "okra.fullname" . }}-manager
{{- end }}

{{- define "okra.editorRoleName" -}}
{{- include "okra.fullname" . }}-editor
{{- end }}

{{- define "okra.viewerRoleName" -}}
{{- include "okra.fullname" . }}-viewer
{{- end }}

{{- define "okra.authProxyServiceName" -}}
{{- include "okra.fullname" . }}-controller-manager-metrics-service
{{- end }}
