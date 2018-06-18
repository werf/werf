module Dapp
  module Kube
    module Dapp
      module Command
        module ChartCreate
          def kube_chart_create
            with_kube_tmp_chart_dir do
              FileUtils.cp_r kube_chart_path, kube_chart_path_for_helm(name) if kube_chart_path.directory? && !options[:force]

              shellout!("helm create #{name}", cwd: kube_chart_path_for_helm)
              kube_create_chart_samples

              FileUtils.rm_rf kube_chart_path
              FileUtils.mv kube_chart_path_for_helm(name), kube_chart_path
            end
          end

          def kube_create_chart_samples
            kube_chart_path_for_helm(name, 'secret-values.yaml').tap { |f| FileUtils.touch(f) unless f.file? }
            kube_chart_path_for_helm(name, kube_chart_secret_dir_name).tap { |dir| FileUtils.mkdir(dir) unless dir.directory? }
            kube_chart_path_for_helm(name, 'templates',  '_envs.tpl').tap do |f|
              f.write begin
                <<-EOF
{{- define "common_envs" }}
- name: A
  value: value
{{-   if eq .Values.global.env "production" }}
- name: B
  value: value
{{-   else }}
- name: B
  value: value2
{{-     if or (eq .Values.global.env "stage") (eq .Values.global.env "test") }}
- name: C
  value: value3
{{-     end }}
{{-   end }}
{{- end }}
                EOF
              end unless f.file?
            end
            kube_chart_path_for_helm(name, 'templates', 'backend.yaml').tap do |f|
              f.write begin
                <<-EOF
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-backend
  labels:
    service: {{ .Chart.Name }}-backend
spec:
  minReadySeconds: 60
  strategy:
    type: RollingUpdate
  replicas: 2
  template:
    metadata:
      labels:
        service: {{ .Chart.Name }}-backend
    spec:
      volumes:
      - name: {{ .Chart.Name }}-backend
        secret:
          secretName: {{ .Chart.Name }}-backend
      containers:
      - command: [ '/bin/bash', '-l', '-c', 'bundle exec ctl start' ]
{{ tuple "dimg-name" . | include "dapp_container_image" | indent 8 }} # or nameless dimg {{ tuple . | include "dapp_container_image" }}
        name: {{ .Chart.Name }}-backend
        livenessProbe:
          httpGet:
            path: /assets/logo.png
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 3
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        env:
{{ tuple "dimg-name" . | include "dapp_container_env" | indent 10 }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  type: ClusterIP
  selector:
    service: {{ .Chart.Name }}-backend
  ports:
  - name: http
    port: 8080
    protocol: TCP
                EOF
              end unless f.file?
            end
          end
        end
      end
    end
  end
end
