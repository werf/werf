module Dapp
  module Kube
    module Kubernetes::Manager
      class Deployment < Base
        def before_deploy
          @revision_before_deploy = 'TODO'
        end

        def watch_till_ready!
          d = nil
          old_d_revision = nil
          dapp.log_process("Replacing kubernetes Deployment #{name}") do
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
          dapp.log_process("Waiting for kubernetes Deployment #{d['metadata']['name']} readiness") do
            known_events_by_pod = {}

            loop do
              d_revision = d.fetch('metadata', {}).fetch('annotations', {}).fetch('deployment.kubernetes.io/revision', nil)

              dapp.log_step("[#{Time.now}] Poll kubernetes Deployment status")
              dapp.with_log_indent do
                dapp.log_info("Target replicas: #{_field_value_for_log(d['spec']['replicas'])}")
                dapp.log_info("Updated replicas: #{_field_value_for_log(d['status']['updatedReplicas'])} / #{_field_value_for_log(d['spec']['replicas'])}")
                dapp.log_info("Available replicas: #{_field_value_for_log(d['status']['availableReplicas'])} / #{_field_value_for_log(d['spec']['replicas'])}")
                dapp.log_info("Ready replicas: #{_field_value_for_log(d['status']['readyReplicas'])} / #{_field_value_for_log(d['spec']['replicas'])}")
                dapp.log_info("Old deployment.kubernetes.io/revision: #{_field_value_for_log(old_d_revision)}")
                dapp.log_info("Current deployment.kubernetes.io/revision: #{_field_value_for_log(d_revision)}")
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

                dapp.with_log_indent do
                  dapp.log_info("Pods:") if rs_pods.any?

                  rs_pods.each do |pod|
                    dapp.with_log_indent do
                      dapp.log_info("* #{pod['metadata']['name']}")

                      known_events_by_pod[pod['metadata']['name']] ||= []
                      pod_events = app.deployment.kubernetes
                        .event_list(fieldSelector: "involvedObject.uid=#{pod['metadata']['uid']}")['items']
                        .reject do |event|
                          known_events_by_pod[pod['metadata']['name']].include? event['metadata']['uid']
                        end

                      if pod_events.any?
                        pod_events.each do |event|
                          dapp.with_log_indent do
                            dapp.log_info("[#{event['metadata']['creationTimestamp']}] #{event['message']}")
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
                        dapp.with_log_indent do
                          dapp.log_warning("Unknown pod readiness condition reason '#{ready_condition['reason']}': #{ready_condition}")
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
          end
        end

      end # Deployment
    end # Kubernetes::Manager
  end # Kube
end # Dapp
