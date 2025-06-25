{{- define "imagespec" }}
{{- $appName := .appName | "defaultapp" }}
clearHistory: true
config:
  labels:
    example.io/app: "{{ $appName }}"
{{- end }}
