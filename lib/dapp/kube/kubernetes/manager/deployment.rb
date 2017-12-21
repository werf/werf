module Dapp
  module Kube
    module Kubernetes::Manager
      class Deployment < Base
        # NOTICE: @revision_before_deploy на данный момент выводится на экран как информация для дебага.
        # NOTICE: deployment.kubernetes.io/revision не меняется при изменении количества реплик, поэтому
        # NOTICE: критерий ожидания по изменению ревизии не верен.
        # NOTICE: Однако, при обновлении deployment ревизия сбрасывается и ожидание переустановки этой ревизии
        # NOTICE: является одним из критериев завершения ожидания на данный момент.

        def before_deploy
          if dapp.kubernetes.deployment? name
            d = Kubernetes::Client::Resource::Deployment.new(dapp.kubernetes.deployment(name))

            @revision_before_deploy = d.annotations['deployment.kubernetes.io/revision']

            unless @revision_before_deploy.nil?
              new_spec = Marshal.load(Marshal.dump(d.spec))
              new_spec.delete('status')
              new_spec.fetch('metadata', {}).fetch('annotations', {}).delete('deployment.kubernetes.io/revision')

              @deployment_before_deploy = Kubernetes::Client::Resource::Deployment.new(dapp.kubernetes.replace_deployment!(name, new_spec))
            end
          end
        end

        def after_deploy
          @deployed_at = Time.now
        end

        def watch_till_ready!
          dapp.log_process("Watch deployment '#{name}' till ready") do
            known_events_by_pod = {}
            known_log_timestamps_by_pod_and_container = {}

            d = @deployment_before_deploy || Kubernetes::Client::Resource::Deployment.new(dapp.kubernetes.deployment(name))

            loop do
              d_revision = d.annotations['deployment.kubernetes.io/revision']

              dapp.log_step("[#{Time.now}] Poll deployment '#{d.name}' status")
              dapp.with_log_indent do
                dapp.log_info("Target replicas: #{_field_value_for_log(d.replicas)}")
                dapp.log_info("Updated replicas: #{_field_value_for_log(d.status['updatedReplicas'])} / #{_field_value_for_log(d.replicas)}")
                dapp.log_info("Available replicas: #{_field_value_for_log(d.status['availableReplicas'])} / #{_field_value_for_log(d.replicas)}")
                dapp.log_info("Ready replicas: #{_field_value_for_log(d.status['readyReplicas'])} / #{_field_value_for_log(d.replicas)}")
                dapp.log_info("Old deployment.kubernetes.io/revision: #{_field_value_for_log(@revision_before_deploy)}")
                dapp.log_info("Current deployment.kubernetes.io/revision: #{_field_value_for_log(d_revision)}")
              end

              rs = nil
              if d_revision
                # Находим актуальный, текущий ReplicaSet.
                # Если такая ситуация, когда есть несколько подходящих по revision ReplicaSet, то берем старейший по дате создания.
                # Также делает kubectl: https://github.com/kubernetes/kubernetes/blob/d86a01570ba243e8d75057415113a0ff4d68c96b/pkg/controller/deployment/util/deployment_util.go#L664
                rs = dapp.kubernetes.replicaset_list['items']
                  .map {|spec| Kubernetes::Client::Resource::Replicaset.new(spec)}
                  .select do |_rs|
                    Array(_rs.metadata['ownerReferences']).any? do |owner_reference|
                      owner_reference['uid'] == d.metadata['uid']
                    end
                  end
                  .select do |_rs|
                    rs_revision = _rs.annotations['deployment.kubernetes.io/revision']
                    (rs_revision and (d_revision == rs_revision))
                  end
                  .sort_by do |_rs|
                    if creation_timestamp = _rs.metadata['creationTimestamp']
                      Time.parse(creation_timestamp)
                    else
                      Time.now
                    end
                  end.first
              end

              if rs
                # Pod'ы связанные с активным ReplicaSet
                rs_pods = dapp.kubernetes.pod_list['items']
                  .map {|spec| Kubernetes::Client::Resource::Pod.new(spec)}
                  .select do |pod|
                    Array(pod.metadata['ownerReferences']).any? do |owner_reference|
                      owner_reference['uid'] == rs.metadata['uid']
                    end
                  end

                dapp.with_log_indent do
                  dapp.log_step("Pods:") if rs_pods.any?

                  rs_pods.each do |pod|
                    dapp.with_log_indent do
                      dapp.log_step(pod.name)

                      known_events_by_pod[pod.name] ||= []
                      pod_events = dapp.kubernetes
                        .event_list(fieldSelector: "involvedObject.uid=#{pod.uid}")['items']
                        .map {|spec| Kubernetes::Client::Resource::Event.new(spec)}
                        .reject do |event|
                          known_events_by_pod[pod.name].include? event.uid
                        end

                      if pod_events.any?
                        dapp.with_log_indent do
                          dapp.log_step("Last events:")
                          pod_events.each do |event|
                            dapp.with_log_indent do
                              dapp.log_info("[#{event.metadata['creationTimestamp']}] #{event.spec['message']}")
                            end
                            known_events_by_pod[pod.name] << event.uid
                          end
                        end
                      end

                      dapp.with_log_indent do
                        pod.containers_names.each do |container_name|
                          known_log_timestamps_by_pod_and_container[pod.name] ||= {}
                          known_log_timestamps_by_pod_and_container[pod.name][container_name] ||= Set.new

                          since_time = nil
                          since_time = @deployed_at.utc.iso8601(9) if @deployed_at

                          log_lines_by_time = []
                          begin
                            log_lines_by_time = dapp.kubernetes.pod_log(pod.name, container: container_name, timestamps: true, sinceTime: since_time)
                              .lines.map(&:strip)
                              .map {|line|
                                timestamp, _, data = line.partition(' ')
                                unless known_log_timestamps_by_pod_and_container[pod.name][container_name].include? timestamp
                                  known_log_timestamps_by_pod_and_container[pod.name][container_name].add timestamp
                                  [timestamp, data]
                                end
                              }.compact
                          rescue Kubernetes::Client::Error::Pod::ContainerCreating, Kubernetes::Client::Error::Pod::PodInitializing
                            next
                          rescue Kubernetes::Client::Error::Base => err
                            dapp.log_warning("#{dapp.log_time}Error while fetching pod's #{pod.name} logs: #{err.message}", stream: dapp.service_stream)
                            next
                          end

                          if log_lines_by_time.any?
                            dapp.log_step("Last container '#{container_name}' log:")
                            dapp.with_log_indent do
                              log_lines_by_time.each do |timestamp, line|
                                dapp.log("[#{timestamp}] #{line}")
                              end
                            end
                          end
                        end
                      end

                      ready_condition = pod.status.fetch('conditions', {}).find {|condition| condition['type'] == 'Ready'}
                      next if (not ready_condition) or (ready_condition['status'] == 'True')

                      if ready_condition['reason'] == 'ContainersNotReady'
                        Array(pod.status['containerStatuses']).each do |container_status|
                          next if container_status['ready']

                          waiting_reason = container_status.fetch('state', {}).fetch('waiting', {}).fetch('reason', nil)
                          case waiting_reason
                          when 'ImagePullBackOff', 'ErrImagePull'
                            raise Kubernetes::Error::Base,
                              code: :image_not_found,
                              data: {pod_name: pod.name,
                                     reason: waiting_reason,
                                     message: container_status['state']['waiting']['message']}
                          when 'CrashLoopBackOff'
                            raise Kubernetes::Error::Base,
                              code: :container_crash,
                              data: {pod_name: pod.name,
                                     reason: waiting_reason,
                                     message: container_status['state']['waiting']['message']}
                          end
                        end
                      else
                        dapp.with_log_indent do
                          dapp.log_warning("#{dapp.log_time}Unknown pod readiness condition reason '#{ready_condition['reason']}': #{ready_condition}", stream: dapp.service_stream)
                        end
                      end
                    end # with_log_indent
                  end # rs_pods.each
                end # with_log_indent
              end

              break if begin
                (d_revision and
                  d.replicas and
                    d.status['updatedReplicas'] and
                      d.status['availableReplicas'] and
                        d.status['readyReplicas'] and
                          (d.status['updatedReplicas'] >= d.replicas) and
                            (d.status['availableReplicas'] >= d.replicas) and
                              (d.status['readyReplicas'] >= d.replicas))
              end

              sleep 5
              d = Kubernetes::Client::Resource::Deployment.new(dapp.kubernetes.deployment(d.name))
            end
          end
        end

        private

        def _field_value_for_log(value)
          value ? value : '-'
        end
      end # Deployment
    end # Kubernetes::Manager
  end # Kube
end # Dapp
