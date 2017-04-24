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

      [:deployment, :service].each do |type|
        define_method "#{type}_exist?" do |name|
          public_send("existing_#{type}s_names").include?(name)
        end

        define_method "existing_#{type}s_names" do
          deployment.kubernetes.public_send(:"#{type}_list", labelSelector: labelSelector)['items'].map do |item|
            item['metadata']['name']
          end
        end

        define_method "replace_#{type}!" do |name, spec|
          hash = send(:"merge_kube_#{type}_spec", deployment.kubernetes.public_send(type, name), spec)
          deployment.kubernetes.public_send(:"replace_#{type}!", name, hash)
        end

        define_method "#{type}_spec_changed?" do |name, spec|
          current_spec = deployment.kubernetes.public_send(type, name)
          current_spec != send(:"merge_kube_#{type}_spec", current_spec, spec)
        end

        [:create, :delete].each do |method|
          define_method "#{method}_#{type}!" do |*args|
            deployment.kubernetes.public_send(:"#{method}_#{type}!", *args)
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

      def update_deployment!(name, conf_spec)
        d = nil
        app.deployment.dapp.log_process("Replacing kubernetes Deployment #{name}") do
          d = app.deployment.kubernetes.replace_deployment!(name, conf_spec)
        end
        _wait_for_deployment(d)
      end

      def _wait_for_deployment(d)
        app.deployment.dapp.log_process("Waiting for kubernetes Deployment #{d['metadata']['name']} readiness", verbose: true) do
          app.deployment.dapp.with_log_indent do
            known_events_by_pod = {}

            loop do
              d_revision = d.fetch('metadata', {}).fetch('annotations', {}).fetch('deployment.kubernetes.io/revision', nil)

              app.deployment.dapp.log_step("[#{Time.now}] Poll kubernetes Deployment status")
              app.deployment.dapp.with_log_indent do
                app.deployment.dapp.log_info("Target replicas: #{d['spec']['replicas']}")
                app.deployment.dapp.log_info("Updated replicas: #{d['status']['updatedReplicas']} / #{d['spec']['replicas']}")
                app.deployment.dapp.log_info("Available replicas: #{d['status']['availableReplicas']} / #{d['spec']['replicas']}")
                app.deployment.dapp.log_info("Ready replicas: #{d['status']['readyReplicas']} / #{d['spec']['replicas']}")
                app.deployment.dapp.log_info("deployment.kubernetes.io/revision: #{d_revision ? d_revision : '-'}")
              end

              rs = nil
              if d_revision
                # Находим актуальный, текущий ReplicaSet.
                # Если такая ситуация, когда есть несколько подходящих по revision ReplicaSet, то берем старейший по дате создания.
                # Также делает kubectl: https://github.com/kubernetes/kubernetes/blob/d86a01570ba243e8d75057415113a0ff4d68c96b/pkg/controller/deployment/util/deployment_util.go#L664
                rs = app.deployment.kubernetes.replicaset_list['items']
                  .select do |_rs|
                    _rs['metadata']['ownerReferences'].any? do |owner_reference|
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
                    pod['metadata']['ownerReferences'].any? do |owner_reference|
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

                      ready_condition = pod['status']['conditions'].find {|condition| condition['type'] == 'Ready'}
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

      def merge_kube_deployment_spec(spec1, spec2)
        merge_kube_controller_spec(spec1, spec2)
      end

      def merge_kube_service_spec(spec1, spec2)
        spec1.kube_in_depth_merge(spec2).tap do |spec|
          spec['spec']['ports'] = begin
            ports1 = spec1['spec']['ports']
            ports2 = spec2['spec']['ports']
            ports2.map do |port2|
              if (port1 = ports1.find { |p| p['name'] == port2['name'] }).nil?
                port2
              else
                port1.kube_in_depth_merge(port2)
              end
            end
          end
        end
      end
    end
  end
end
