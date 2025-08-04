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
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "smart-garden-bot.selectorLabels" -}}
app.kubernetes.io/name: {{ include "smart-garden-bot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "smart-garden-bot.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "smart-garden-bot.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the operator service account to use
*/}}
{{- define "smart-garden-bot.operatorServiceAccountName" -}}
{{- if .Values.operator.serviceAccount.create }}
{{- default (printf "%s-operator" (include "smart-garden-bot.fullname" .)) .Values.operator.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.operator.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create a default fully qualified PostgreSQL name.
*/}}
{{- define "smart-garden-bot.postgresql.fullname" -}}
{{- include "common.names.dependency.fullname" (dict "chartName" "postgresql" "chartValues" .Values.postgresql "context" $) -}}
{{- end }}

{{/*
Create a default fully qualified Redis name.
*/}}
{{- define "smart-garden-bot.redis.fullname" -}}
{{- include "common.names.dependency.fullname" (dict "chartName" "redis" "chartValues" .Values.redis "context" $) -}}
{{- end }}

{{/*
Get the PostgreSQL secret name
*/}}
{{- define "smart-garden-bot.postgresql.secretName" -}}
{{- if .Values.postgresql.auth.existingSecret }}
{{- .Values.postgresql.auth.existingSecret }}
{{- else }}
{{- include "smart-garden-bot.postgresql.fullname" . }}
{{- end }}
{{- end }}

{{/*
Get the Redis secret name
*/}}
{{- define "smart-garden-bot.redis.secretName" -}}
{{- if .Values.redis.auth.existingSecret }}
{{- .Values.redis.auth.existingSecret }}
{{- else }}
{{- include "smart-garden-bot.redis.fullname" . }}
{{- end }}
{{- end }}

{{/*
Validate required values
*/}}
{{- define "smart-garden-bot.validateValues" -}}
{{- if and .Values.webapp.enabled (not .Values.webapp.env.AUTH0_BASE_URL) }}
{{- fail "webapp.env.AUTH0_BASE_URL is required when webapp is enabled" }}
{{- end }}
{{- if and .Values.apiserver.enabled (not .Values.apiserver.env.DB_HOST) }}
{{- fail "apiserver.env.DB_HOST is required when apiserver is enabled" }}
{{- end }}
{{- if and .Values.operator.enabled (not .Values.operator.rbac.create) }}
{{- fail "operator.rbac.create must be true when operator is enabled" }}
{{- end }}
{{- end }}

{{/*
Generate certificates for webhook
*/}}
{{- define "smart-garden-bot.webhook.certs" -}}
{{- $altNames := list ( printf "%s-operator-webhook-service" (include "smart-garden-bot.fullname" .) ) ( printf "%s-operator-webhook-service.%s" (include "smart-garden-bot.fullname" .) .Release.Namespace ) ( printf "%s-operator-webhook-service.%s.svc" (include "smart-garden-bot.fullname" .) .Release.Namespace ) -}}
{{- $ca := genCA "smart-garden-bot-ca" 365 -}}
{{- $cert := genSignedCert ( printf "%s-operator-webhook-service" (include "smart-garden-bot.fullname" .) ) nil $altNames 365 $ca -}}
tls.crt: {{ $cert.Cert | b64enc }}
tls.key: {{ $cert.Key | b64enc }}
ca.crt: {{ $ca.Cert | b64enc }}
{{- end }}

{{/*
Common environment variables for all components
*/}}
{{- define "smart-garden-bot.commonEnv" -}}
- name: RELEASE_NAME
  value: {{ .Release.Name | quote }}
- name: RELEASE_NAMESPACE
  value: {{ .Release.Namespace | quote }}
- name: CHART_NAME
  value: {{ .Chart.Name | quote }}
- name: CHART_VERSION
  value: {{ .Chart.Version | quote }}
{{- end }}

{{/*
Database connection URL
*/}}
{{- define "smart-garden-bot.databaseUrl" -}}
{{- if .Values.postgresql.enabled }}
postgresql://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ include "smart-garden-bot.postgresql.fullname" . }}:{{ .Values.postgresql.primary.service.ports.postgresql }}/{{ .Values.postgresql.auth.database }}
{{- else if .Values.cloudnative-pg.enabled }}
postgresql://postgres:$(DB_PASSWORD)@{{ .Values.postgresqlCluster.name }}-rw:5432/{{ .Values.postgresqlCluster.spec.bootstrap.initdb.database }}
{{- else }}
postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)
{{- end }}
{{- end }}

{{/*
Redis connection URL
*/}}
{{- define "smart-garden-bot.redisUrl" -}}
{{- if .Values.redis.enabled }}
{{- if .Values.redis.auth.enabled }}
redis://:$(REDIS_PASSWORD)@{{ include "smart-garden-bot.redis.fullname" . }}-master:{{ .Values.redis.master.service.ports.redis }}
{{- else }}
redis://{{ include "smart-garden-bot.redis.fullname" . }}-master:{{ .Values.redis.master.service.ports.redis }}
{{- end }}
{{- else }}
redis://$(REDIS_HOST):$(REDIS_PORT)
{{- end }}
{{- end }}

{{/*
Generate API server configuration
*/}}
{{- define "smart-garden-bot.apiserver.config" -}}
database:
  url: {{ include "smart-garden-bot.databaseUrl" . }}
  maxConnections: 25
  minConnections: 5
  
redis:
  url: {{ include "smart-garden-bot.redisUrl" . }}
  
server:
  port: {{ .Values.apiserver.service.targetPort }}
  readTimeout: 30s
  writeTimeout: 30s
  
logging:
  level: {{ .Values.apiserver.env.LOG_LEVEL | default "info" }}
  format: json
  
metrics:
  enabled: {{ .Values.apiserver.env.METRICS_ENABLED | default "true" }}
  port: 8080
  
cors:
  allowedOrigins:
    {{- range .Values.config.cors.allowedOrigins }}
    - {{ . | quote }}
    {{- end }}
  allowedMethods:
    {{- range .Values.config.cors.allowedMethods }}
    - {{ . | quote }}
    {{- end }}
  allowedHeaders:
    {{- range .Values.config.cors.allowedHeaders }}
    - {{ . | quote }}
    {{- end }}
{{- end }}

{{/*
Generate webapp configuration
*/}}
{{- define "smart-garden-bot.webapp.config" -}}
api:
  baseUrl: {{ .Values.webapp.env.NEXT_PUBLIC_API_BASE_URL }}
  timeout: 30000
  
auth0:
  domain: {{ .Values.webapp.env.AUTH0_ISSUER_BASE_URL | replace "https://" "" }}
  clientId: $(AUTH0_CLIENT_ID)
  clientSecret: $(AUTH0_CLIENT_SECRET)
  baseUrl: {{ .Values.webapp.env.AUTH0_BASE_URL }}
  secret: $(AUTH0_SECRET)
  
features:
  {{- toYaml .Values.config.features | nindent 2 }}
{{- end }}