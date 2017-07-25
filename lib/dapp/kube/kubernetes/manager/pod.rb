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

        def wait_till_running!
          loop do
            pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(name))
            break if pod.phase == 'Running'
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
