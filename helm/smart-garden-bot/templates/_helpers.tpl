{{/*
Expand the name of the chart.
*/}}
{{- define "smart-garden-bot.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "smart-garden-bot.fullname" -}}
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
{{- define "smart-garden-bot.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "smart-garden-bot.labels" -}}
helm.sh/chart: {{ include "smart-garden-bot.chart" . }}
{{ include "smart-garden-bot.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Web app labels
*/}}
{{- define "smart-garden-bot.labels.web" -}}
helm.sh/chart: {{ include "smart-garden-bot.chart" . }}
{{ include "smart-garden-bot.selectorLabels.web" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
API server labels
*/}}
{{- define "smart-garden-bot.labels.api" -}}
helm.sh/chart: {{ include "smart-garden-bot.chart" . }}
{{ include "smart-garden-bot.selectorLabels.api" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Operator labels
*/}}
{{- define "smart-garden-bot.labels.operator" -}}
helm.sh/chart: {{ include "smart-garden-bot.chart" . }}
{{ include "smart-garden-bot.selectorLabels.operator" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "smart-garden-bot.selectorLabels" -}}
app.kubernetes.io/name: {{ include "smart-garden-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Web app selector labels
*/}}
{{- define "smart-garden-bot.selectorLabels.web" -}}
app.kubernetes.io/name: {{ include "smart-garden-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: web
{{- end }}

{{/*
API server selector labels
*/}}
{{- define "smart-garden-bot.selectorLabels.api" -}}
app.kubernetes.io/name: {{ include "smart-garden-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: api
{{- end }}

{{/*
Operator selector labels
*/}}
{{- define "smart-garden-bot.selectorLabels.operator" -}}
app.kubernetes.io/name: {{ include "smart-garden-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: operator
{{- end }}

{{/*
Create the name of the service account to use for web app
*/}}
{{- define "smart-garden-bot.serviceAccountName.web" -}}
{{- if .Values.webApp.serviceAccount.create }}
{{- default (printf "%s-web" (include "smart-garden-bot.fullname" .)) .Values.webApp.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.webApp.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for API server
*/}}
{{- define "smart-garden-bot.serviceAccountName.api" -}}
{{- if .Values.apiServer.serviceAccount.create }}
{{- default (printf "%s-api" (include "smart-garden-bot.fullname" .)) .Values.apiServer.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.apiServer.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for operator
*/}}
{{- define "smart-garden-bot.serviceAccountName.operator" -}}
{{- if .Values.operator.serviceAccount.create }}
{{- default (printf "%s-operator" (include "smart-garden-bot.fullname" .)) .Values.operator.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.operator.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
PostgreSQL host
*/}}
{{- define "smart-garden-bot.postgresql.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" .Release.Name }}
{{- else }}
{{- .Values.externalDatabase.host }}
{{- end }}
{{- end }}

{{/*
PostgreSQL port
*/}}
{{- define "smart-garden-bot.postgresql.port" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.service.ports.postgresql | default 5432 }}
{{- else }}
{{- .Values.externalDatabase.port | default 5432 }}
{{- end }}
{{- end }}

{{/*
PostgreSQL database name
*/}}
{{- define "smart-garden-bot.postgresql.database" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.externalDatabase.database }}
{{- end }}
{{- end }}

{{/*
PostgreSQL username
*/}}
{{- define "smart-garden-bot.postgresql.username" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.username }}
{{- else }}
{{- .Values.externalDatabase.username }}
{{- end }}
{{- end }}

{{/*
PostgreSQL secret name
*/}}
{{- define "smart-garden-bot.postgresql.secretName" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" .Release.Name }}
{{- else }}
{{- .Values.externalDatabase.existingSecret }}
{{- end }}
{{- end }}

{{/*
PostgreSQL secret key
*/}}
{{- define "smart-garden-bot.postgresql.secretPasswordKey" -}}
{{- if .Values.postgresql.enabled }}
{{- "password" }}
{{- else }}
{{- .Values.externalDatabase.existingSecretPasswordKey }}
{{- end }}
{{- end }}

{{/*
Redis host
*/}}
{{- define "smart-garden-bot.redis.host" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master" .Release.Name }}
{{- else }}
{{- .Values.externalRedis.host }}
{{- end }}
{{- end }}

{{/*
Redis port
*/}}
{{- define "smart-garden-bot.redis.port" -}}
{{- if .Values.redis.enabled }}
{{- .Values.redis.master.service.ports.redis | default 6379 }}
{{- else }}
{{- .Values.externalRedis.port | default 6379 }}
{{- end }}
{{- end }}

{{/*
Redis secret name
*/}}
{{- define "smart-garden-bot.redis.secretName" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis" .Release.Name }}
{{- else }}
{{- .Values.externalRedis.existingSecret }}
{{- end }}
{{- end }}

{{/*
Redis secret key
*/}}
{{- define "smart-garden-bot.redis.secretPasswordKey" -}}
{{- if .Values.redis.enabled }}
{{- "redis-password" }}
{{- else }}
{{- .Values.externalRedis.existingSecretPasswordKey }}
{{- end }}
{{- end }}