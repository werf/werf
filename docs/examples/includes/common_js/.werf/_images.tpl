{{- define "image-backend" }}
image: backend
dockerfile: backend.Dockerfile
{{- end }}

{{- define "image-frontend" }}
{{- $backendUrl := default "http://example.org" .backendUrl }}
image: frontend
from: nginx:alpine
dockerfile: frontend.Dockerfile
args:
  BACKEND_URL: "{{ $backendUrl }}"
{{- end }}
