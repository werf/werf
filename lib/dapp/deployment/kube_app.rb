module Dapp
  module Deployment
    class KubeApp < KubeBase
      module Error
        class Base < ::Dapp::Deployment::Error::Deployment
          def initialize(**net_status)
            super(**net_status, context: :kube_app)
          end
        end
      end

      attr_reader :app

      def initialize(app)
        @app = app
      end

      def deployment
        app.deployment
      end

      def labels
        deployment.kube.labels.merge('dapp-app' => app.name)
      end

      [[:deployment, 'Deployment'], [:service, 'Service']].each do |(type, resource_name)|
        define_method "#{type}_exist?" do |name|
          public_send("existing_#{type}s_names").include?(name)
        end

        define_method "existing_#{type}s_names" do
          deployment.kubernetes.public_send(:"#{type}_list", labelSelector: labelSelector)['items'].map do |item|
            item['metadata']['name']
          end
        end

        define_method "#{type}_spec_changed?" do |name, spec|
          current_spec = deployment.kubernetes.public_send(type, name)
          current_spec != send(:"merge_kube_#{type}_spec", current_spec, spec)
        end

        define_method "delete_#{type}!" do |*args|
          deployment.kubernetes.public_send(:"delete_#{type}!", *args)
        end

        define_method "apply_#{type}!" do |name, spec|
          if app.kube.send(:"#{type}_exist?", name)
            if app.kube.send(:"#{type}_spec_changed?", name, spec)
              app.kube.send(:"replace_#{type}!", name, spec)
            else
              app.deployment.dapp.log_step("Kubernetes #{resource_name} #{name} is up to date")
            end
          else
            app.kube.send(:"create_#{type}!", spec)
          end
        end
      end

      def create_service!(conf_spec)
        srv = nil
        app.deployment.dapp.log_process("Creating kubernetes Service #{conf_spec['metadata']['name']}") do
          srv = app.deployment.kubernetes.create_service!(conf_spec)
        end
        _dump_service_info srv
      end

      def replace_service!(name, conf_spec)
        srv = nil
        app.deployment.dapp.log_process("Replacing kubernetes Service #{name}") do
          old_spec = deployment.kubernetes.service(name)
          new_spec = merge_kube_service_spec(old_spec, conf_spec)
          srv = app.deployment.kubernetes.replace_service!(name, new_spec)
        end
        _dump_service_info srv
      end

      def _dump_service_info(srv)
        app.deployment.dapp.with_log_indent do
          app.deployment.dapp.log_info("type: #{srv['spec']['type']}")
          app.deployment.dapp.log_info("clusterIP: #{srv['spec']['clusterIP']}")

          srv['spec'].fetch('ports', []).each do |port|
            app.deployment.dapp.log_info("Port #{port['port']}:")
            app.deployment.dapp.with_log_indent do
              %w(protocol targetPort nodePort).each do |field_name|
                app.deployment.dapp.log_info("#{field_name}: #{_field_value_for_log(port[field_name])}")
              end
            end
          end
        end
      end

      def create_deployment!(conf_spec)
        d = nil
        app.deployment.dapp.log_process("Creating kubernetes Deployment #{conf_spec['metadata']['name']}") do
          d = app.deployment.kubernetes.create_deployment!(conf_spec)
        end
        _wait_for_deployment(d)
      end

      def replace_deployment!(name, conf_spec)
        d = nil
        old_d_revision = nil
        app.deployment.dapp.log_process("Replacing kubernetes Deployment #{name}") do
          old_spec = deployment.kubernetes.deployment(name)
          old_d_revision = old_spec.fetch('metadata', {}).fetch('annotations', {}).fetch('deployment.kubernetes.io/revision', nil)
          new_spec = merge_kube_deployment_spec(old_spec, conf_spec)
          new_spec.fetch('metadata', {}).fetch('annotations', {}).delete('deployment.kubernetes.io/revision')
          d = app.deployment.kubernetes.replace_deployment!(name, new_spec)
        end
        _wait_for_deployment(d, old_d_revision: old_d_revision)
      end

      # NOTICE: old_d_revision на данный момент выводится на экран как информация для дебага.
      # NOTICE: deployment.kubernetes.io/revision не меняется при изменении количества реплик, поэтому
      # NOTICE: критерий ожидания по изменению ревизии не верен.
      # NOTICE: Однако, при обновлении deployment ревизия сбрасывается и ожидание переустановки этой ревизии
      # NOTICE: является одним из критериев завершения ожидания на данный момент.
      def _wait_for_deployment(d, old_d_revision: nil)
        app.deployment.dapp.log_process("Waiting for kubernetes Deployment #{d['metadata']['name']} readiness") do
          app.deployment.dapp.with_log_indent do
            known_events_by_pod = {}

            loop do
              d_revision = d.fetch('metadata', {}).fetch('annotations', {}).fetch('deployment.kubernetes.io/revision', nil)

              app.deployment.dapp.log_step("[#{Time.now}] Poll kubernetes Deployment status")
              app.deployment.dapp.with_log_indent do
                app.deployment.dapp.log_info("Target replicas: #{_field_value_for_log(d['spec']['replicas'])}")
                app.deployment.dapp.log_info("Updated replicas: #{_field_value_for_log(d['status']['updatedReplicas'])} / #{_field_value_for_log(d['spec']['replicas'])}")
                app.deployment.dapp.log_info("Available replicas: #{_field_value_for_log(d['status']['availableReplicas'])} / #{_field_value_for_log(d['spec']['replicas'])}")
                app.deployment.dapp.log_info("Ready replicas: #{_field_value_for_log(d['status']['readyReplicas'])} / #{_field_value_for_log(d['spec']['replicas'])}")
                app.deployment.dapp.log_info("Old deployment.kubernetes.io/revision: #{_field_value_for_log(old_d_revision)}")
                app.deployment.dapp.log_info("Current deployment.kubernetes.io/revision: #{_field_value_for_log(d_revision)}")
              end

              rs = nil
              if d_revision
                # Находим актуальный, текущий ReplicaSet.
                # Если такая ситуация, когда есть несколько подходящих по revision ReplicaSet, то берем старейший по дате создания.
                # Также делает kubectl: https://github.com/kubernetes/kubernetes/blob/d86a01570ba243e8d75057415113a0ff4d68c96b/pkg/controller/deployment/util/deployment_util.go#L664
                rs = app.deployment.kubernetes.replicaset_list['items']
                  .select do |_rs|
                    Array(_rs['metadata']['ownerReferences']).any? do |owner_reference|
                      owner_reference['uid'] == d['metadata']['uid']
                    end
                  end
                  .select do |_rs|
                    rs_revision = _rs.fetch('metadata', {}).fetch('annotations', {}).fetch('deployment.kubernetes.io/revision', nil)
                    (rs_revision and (d_revision == rs_revision))
                  end
                  .sort_by do |_rs|
                    Time.parse _rs['metadata']['creationTimestamp']
                  end.first
              end

              if rs
                # Pod'ы связанные с активным ReplicaSet
                rs_pods = app.deployment.kubernetes
                  .pod_list(labelSelector: labels.map{|k, v| "#{k}=#{v}"}.join(','))['items']
                  .select do |pod|
                    Array(pod['metadata']['ownerReferences']).any? do |owner_reference|
                      owner_reference['uid'] == rs['metadata']['uid']
                    end
                  end

                app.deployment.dapp.with_log_indent do
                  app.deployment.dapp.log_info("Pods:") if rs_pods.any?

                  rs_pods.each do |pod|
                    app.deployment.dapp.with_log_indent do
                      app.deployment.dapp.log_info("* #{pod['metadata']['name']}")

                      known_events_by_pod[pod['metadata']['name']] ||= []
                      pod_events = app.deployment.kubernetes
                        .event_list(fieldSelector: "involvedObject.uid=#{pod['metadata']['uid']}")['items']
                        .reject do |event|
                          known_events_by_pod[pod['metadata']['name']].include? event['metadata']['uid']
                        end

                      if pod_events.any?
                        pod_events.each do |event|
                          app.deployment.dapp.with_log_indent do
                            app.deployment.dapp.log_info("[#{event['metadata']['creationTimestamp']}] #{event['message']}")
                          end
                          known_events_by_pod[pod['metadata']['name']] << event['metadata']['uid']
                        end
                      end

                      ready_condition = pod['status'].fetch('conditions', {}).find {|condition| condition['type'] == 'Ready'}
                      next if (not ready_condition) or (ready_condition['status'] == 'True')

                      if ready_condition['reason'] == 'ContainersNotReady'
                        Array(pod['status']['containerStatuses']).each do |container_status|
                          next if container_status['ready']

                          waiting_reason = container_status.fetch('state', {}).fetch('waiting', {}).fetch('reason', nil)
                          case waiting_reason
                          when 'ImagePullBackOff', 'ErrImagePull'
                            raise Error::Base,
                              code: :image_not_found,
                              data: {app: app.name,
                                     pod_name: pod['metadata']['name'],
                                     reason: container_status['state']['waiting']['reason'],
                                     message: container_status['state']['waiting']['message']}
                          when 'CrashLoopBackOff'
                            raise Error::Base,
                              code: :container_crash,
                              data: {app: app.name,
                                     pod_name: pod['metadata']['name'],
                                     reason: container_status['state']['waiting']['reason'],
                                     message: container_status['state']['waiting']['message']}
                          end
                        end
                      else
                        app.deployment.dapp.with_log_indent do
                          app.deployment.dapp.log_warning("Unknown pod readiness condition reason '#{ready_condition['reason']}': #{ready_condition}")
                        end
                      end
                    end # with_log_indent
                  end # rs_pods.each
                end # with_log_indent
              end

              break if begin
                d_revision and
                  d['spec']['replicas'] and
                    d['status']['updatedReplicas'] and
                      d['status']['availableReplicas'] and
                        d['status']['readyReplicas'] and
                          (d['status']['updatedReplicas'] >= d['spec']['replicas']) and
                            (d['status']['availableReplicas'] >= d['spec']['replicas']) and
                              (d['status']['readyReplicas'] >= d['spec']['replicas'])
              end

              sleep 1
              d = app.deployment.kubernetes.deployment(d['metadata']['name'])
            end
          end # with_log_indent
        end
      end

      def _field_value_for_log(value)
        value ? value : '-'
      end

      def merge_kube_deployment_spec(spec1, spec2)
        merge_kube_controller_spec(spec1, spec2)
      end

      def merge_kube_service_spec(spec1, spec2)
        spec1.kube_in_depth_merge(spec2).tap do |spec|
          spec['metadata'] ||= {}
          metadata_labels = spec2.fetch('metadata', {}).fetch('labels', nil)
          spec['metadata']['labels'] = metadata_labels if metadata_labels

          spec['spec'] ||= {}
          spec_selector = spec2.fetch('spec', {}).fetch('selector', nil)
          spec['spec']['selector'] = spec_selector if spec_selector
          spec['spec']['ports'] = begin
            ports1 = spec1.fetch('spec', {}).fetch('ports', [])
            ports2 = spec2.fetch('spec', {}).fetch('ports', [])
            ports2.map do |port2|
              if (port1 = ports1.find { |p| p['name'] == port2['name'] }).nil?
                port2
              else
                port = port1.merge(port2)
                port.delete('nodePort') if spec['spec']['type'] == 'ClusterIP'
                port
              end
            end
          end
        end
      end
    end
  end
end
