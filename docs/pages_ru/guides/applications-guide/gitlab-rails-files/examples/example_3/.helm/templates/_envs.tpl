{{- define "apps_envs" }}
- name: RAILS_MASTER_KEY
  value: {{ .Values.rails.master_key}}
- name: RAILS_ENV
  value: production
- name: RAILS_LOG_TO_STDOUT
  value: "true"
{{- end }}

{{- define "database_envs" }}
- name: DATABASE_HOST
  value: {{ .Chart.Name }}-{{ .Values.global.env }}-postgresql
- name: DATABASE_NAME
  value: {{ .Values.postgresql.postgresqlDatabase }}
- name: DATABASE_USERNAME
  value: {{ .Values.postgresql.postgresqlUsername }}
- name: DATABASE_PASSWORD
  value: {{ .Values.postgresql.postgresqlPassword }}
{{- end }}
