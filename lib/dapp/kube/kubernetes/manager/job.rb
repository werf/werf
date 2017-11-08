module Dapp
  module Kube
    module Kubernetes::Manager
      class Job < Base
        def initialize(dapp, name)
          super(dapp, name)

          @processed_pods_names = []
        end

        def wait_till_exists!
          loop do
            break if dapp.kubernetes.job?(name)
            sleep 0.1
          end
        end

        def watch_till_done!
          wait_till_exists!

          job = Kubernetes::Client::Resource::Job.new(dapp.kubernetes.job(name))

          loop do
            # Получить очередной pod для обработки
            process_pod = dapp.kubernetes.pod_list.fetch('items', [])
              .select do |pod_spec|
                pod_spec.fetch('metadata', {}).fetch('labels', {})['controller-uid'] == job.uid
              end
              .reject do |pod_spec|
                @processed_pods_names.include? pod_spec.fetch('metadata', {})['name']
              end
              .sort_by do |pod_spec|
                Time.parse(pod_spec.fetch('metadata', {})['creationTimestamp'])
              end
              .map {|pod_spec| Kubernetes::Client::Resource::Pod.new(pod_spec)}
              .first

            if process_pod.nil?
              job = Kubernetes::Client::Resource::Job.new(dapp.kubernetes.job(name))
              if job.succeeded?
                break
              elsif job.failed?
                dapp.log_warning "#{dapp.log_time}Job '#{name}' has been failed: #{job.spec['status']}", stream: dapp.service_stream
                break
              end

              sleep 0.1

              next
            end

            pod_manager = Kubernetes::Manager::Pod.new(dapp, process_pod.name)
            begin
              pod_manager.watch_till_done!
            rescue Kubernetes::Client::Error::Pod::NotFound => err
              dapp.log_warning "#{dapp.log_time}Pod '#{pod_manager.name}' has been deleted", stream: dapp.service_stream
            ensure
              @processed_pods_names << process_pod.name
            end
          end
        end
      end # Job
    end # Kubernetes::Manager
  end # Kube
end # Dapp
