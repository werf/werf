module Dapp
  module Kube
    module Kubernetes::Manager
      class Pod < Base
        def containers
          pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(name))
          @containers ||= pod.containers_names.map do |container_name|
            Container.new(dapp, container_name, self)
          end
        end

        def check_readiness_condition_if_available!(pod)
          return if [nil, "True"].include? pod.ready_condition_status

          if pod.ready_condition['reason'] == 'ContainersNotReady'
            [*Array(pod.status["initContainerStatuses"]), *Array(pod.status["containerStatuses"])].each do |container_status|
              next if container_status['ready']

              waiting_reason = container_status.fetch('state', {}).fetch('waiting', {}).fetch('reason', nil)
              case waiting_reason
              when 'ImagePullBackOff', 'ErrImagePull'
                raise Kubernetes::Error::Default,
                  code: :bad_image,
                  data: {pod_name: pod.name,
                         reason: waiting_reason,
                         message: container_status['state']['waiting']['message']}
              when 'CrashLoopBackOff'
                raise Kubernetes::Error::Default,
                  code: :container_crash,
                  data: {pod_name: pod.name,
                         reason: waiting_reason,
                         message: container_status['state']['waiting']['message']}
              end
            end
          else
            dapp.with_log_indent do
              dapp.log_warning("#{dapp.log_time}Unknown pod readiness condition reason '#{pod.ready_condition['reason']}': #{pod.ready_condition}", stream: dapp.service_stream)
            end
          end
        end

        def wait_till_launched!
          loop do
            pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(name))

            break if pod.phase != "Pending"

            check_readiness_condition_if_available!(pod)

            sleep 0.1
          end
        end

        def watch_till_done!
          process_queue = containers.map {|c| [c, nil]}
          loop do
            container, last_processed_at = process_queue.shift
            break unless container

            sleep 1 if last_processed_at and (Time.now - last_processed_at < 1)

            container.watch_till_terminated!
            process_queue.push([container, Time.now]) unless container.done?
          end
        end
      end # Pod
    end # Kubernetes::Manager
  end # Kube
end # Dapp
