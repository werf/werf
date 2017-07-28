module Dapp
  module Kube
    module Dapp
      module Command
        module ChartCreate
          def kube_chart_create
            with_kube_tmp_chart_dir do
              FileUtils.cp_r kube_chart_path, kube_tmp_chart_path(name) if kube_chart_path.directory? && !options[:force]

              shellout!("helm create #{name}", cwd: kube_tmp_chart_path)
              kube_create_chart_samples

              FileUtils.rm_rf kube_chart_path
              FileUtils.mv kube_tmp_chart_path(name), kube_chart_path
            end
          end

          def kube_create_chart_samples
            kube_tmp_chart_path(name, 'secret-values.yaml').tap { |f| FileUtils.touch(f) unless f.file? }
            kube_tmp_chart_path(name, kube_chart_secret_dir_name).tap { |dir| FileUtils.mkdir(dir) unless dir.directory? }
            kube_tmp_chart_path(name, 'templates',  '_envs.tpl').tap do |f|
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
{{-     if or (eq .Values.global.env "staging") (eq .Values.global.env "testing") }}
- name: C
  value: value3
{{-     end }}
{{-   end }}
{{- end }}
                EOF
              end unless f.file?
            end
            kube_tmp_chart_path(name, 'templates', 'backend.yaml').tap do |f|
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
        image: {{ tuple . 'specific_name' | include "dimg" }} # or nameless dimg {{ tuple . | include "dimg" }}
        imagePullPolicy: Always
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
{{- include "common_envs" . | indent 8 }}
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