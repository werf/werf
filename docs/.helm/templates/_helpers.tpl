{{- define "resources" }}
resources:
  requests:
    memory: {{ pluck .Values.werf.env .Values.resources.requests.memory | first | default .Values.resources.requests.memory._default }}
  limits:
    memory: {{ pluck .Values.werf.env .Values.resources.requests.memory | first | default .Values.resources.requests.memory._default }}
{{- end }}

{{- define "targetCluster" }}
{{- if eq .Values.werf.env "production" }}
{{- $targetCluster := .Values.global.targetCluster | default "" -}}
{{- if or (eq $targetCluster "eu") (eq $targetCluster "ru") -}}
{{- $targetCluster -}}
{{- else -}}
{{- fail "For production, set global.targetCluster to either 'eu' or 'ru'." -}}
{{- end -}}
{{- else -}}
eu
{{- end }}
{{- end }}

{{- define "ingressClassName" }}
{{- pluck .Values.werf.env .Values.ingressClassName | first | default .Values.ingressClassName._default -}}
{{- end }}


{{- define "clusterPlacement" }}
{{- $targetCluster := include "targetCluster" . -}}
{{- $clusterConfig := get (.Values.clusters | default dict) $targetCluster | default dict -}}
{{- $placement := get $clusterConfig "placement" | default dict -}}
{{- with (get $placement "nodeSelector") }}
nodeSelector:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with (get $placement "tolerations") }}
tolerations:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with (get $placement "affinity") }}
affinity:
{{ toYaml . | indent 2 }}
{{- end }}
{{- end }}
