#/*
# * Copyright 2019 Intel Corporation, Inc
# *
# * Licensed under the Apache License, Version 2.0 (the "License");
# * you may not use this file except in compliance with the License.
# * You may obtain a copy of the License at
# *
# *     http://www.apache.org/licenses/LICENSE-2.0
# *
# * Unless required by applicable law or agreed to in writing, software
# * distributed under the License is distributed on an "AS IS" BASIS,
# * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# * See the License for the specific language governing permissions and
# * limitations under the License.
# */


Expand the name of the chart.
*/}}
{{- define "metallb.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "metallb.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "metallb.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the controller service account to use
*/}}
{{- define "metallb.controllerServiceAccountName" -}}
{{- if .Values.serviceAccounts.controller.create -}}
    {{ default (printf "%s-controller" (include "metallb.fullname" .)) .Values.serviceAccounts.controller.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.controller.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the speaker service account to use
*/}}
{{- define "metallb.speakerServiceAccountName" -}}
{{- if .Values.serviceAccounts.speaker.create -}}
    {{ default (printf "%s-speaker" (include "metallb.fullname" .)) .Values.serviceAccounts.speaker.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.speaker.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the settings ConfigMap to use.
*/}}
{{- define "metallb.configMapName" -}}
{{- if .Values.config -}}
    {{ include "metallb.fullname" . }}
{{- else -}}
    {{ .Values.existingConfigMap }}
{{- end -}}
{{- end -}}
