{{- define "imagespec" }}
{{- $appName := default "defaultapp" .appName}}
clearHistory: true
config:
  labels:
    example.io/app: "{{ $appName }}"
{{- end }}
