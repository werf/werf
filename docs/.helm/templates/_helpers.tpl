{{- define "resources" }}
resources:
  requests:
    memory: {{ pluck .Values.werf.env .Values.resources.requests.memory | first | default .Values.resources.requests.memory._default }}
  limits:
    memory: {{ pluck .Values.werf.env .Values.resources.requests.memory | first | default .Values.resources.requests.memory._default }}
{{- end }}