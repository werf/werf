{{- define "image-backend" }}
image: backend
dockerfile: backend.Dockerfile
{{- end }}

{{- define "image-frontend" }}
{{- $backendUrl := default "http://example.org" .backendUrl }}
image: frontend
dockerfile: frontend.Dockerfile
args:
  BACKEND_URL: "{{ $backendUrl }}"
{{- end }}
