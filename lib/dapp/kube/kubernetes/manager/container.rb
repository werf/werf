module Dapp
  module Kube
    module Kubernetes::Manager
      class Container < Base
        attr_reader :pod_manager

        def initialize(dapp, name, pod_manager)
          super(dapp, name)

          @pod_manager = pod_manager
          @processed_containers_ids = []
          @processed_log_till_time = nil
          @processed_log_timestamps = Set.new
        end

        def watch_till_terminated!
          pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(pod_manager.name))
          _, container_state_data = pod.container_state(name)
          return if @processed_containers_ids.include? container_state_data['containerID']

          pod_manager.wait_till_launched!

          pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(pod_manager.name))
          container_state, container_state_data = pod.container_state(name)

          dapp.log_process("Watch pod's '#{pod_manager.name}' container '#{name}' log") do
            loop do
              pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(pod_manager.name))
              container_state, container_state_data = pod.container_state(name)

              if container_state == "waiting"
                if container_state_data["reason"] == "RunContainerError"
                  raise Kubernetes::Error::Default, code: :container_stuck, data: {
                    state_reason: container_state_data["reason"],
                    state_message: container_state_data["message"],
                    state: container_state,
                    pod_name: pod_manager.name,
                    container_name: name
                  }
                else
                  sleep 0.1
                  next
                end
              end

              chunk_lines_by_time = {}
              begin
                chunk_lines_by_time = dapp.kubernetes.pod_log(pod_manager.name, container: name, timestamps: true, sinceTime: @processed_log_till_time)
                  .lines
                  .map(&:strip)
                  .map do |line|
                    timestamp, _, data = line.partition(' ')
                    [timestamp, data]
                  end
                  .reject {|timestamp, _| @processed_log_timestamps.include? timestamp}
              rescue Kubernetes::Client::Error::Pod::ContainerCreating, Kubernetes::Client::Error::Pod::PodInitializing
                sleep 0.1
                next
              rescue Kubernetes::Client::Error::Base => err
                dapp.log_warning("#{dapp.log_time}Error while fetching pod's #{pod_manager.name} logs: #{err.message}", stream: dapp.service_stream)
                break
              end

              chunk_lines_by_time.each do |timestamp, data|
                dapp.log("[#{timestamp}] #{data}")
                @processed_log_timestamps.add timestamp
              end

              if container_state == 'terminated'
                failed = (container_state_data['exitCode'].to_i != 0)

                dapp.log_warning("".tap do |msg|
                  msg << "Pod's '#{pod_manager.name}' container '#{name}' has been terminated unsuccessfuly: "
                  msg << container_state_data.to_s
                end) if failed

                @processed_containers_ids << container_state_data['containerID']

                break
              end

              @processed_log_till_time = chunk_lines_by_time.last.first if chunk_lines_by_time.any?

              sleep 0.1 if chunk_lines_by_time.empty?
            end
          end # log_process
        end

        def done?
          pod = Kubernetes::Client::Resource::Pod.new(dapp.kubernetes.pod(pod_manager.name))
          container_state, container_state_data = pod.container_state(name)
          if container_state == 'terminated'
            failed = (container_state_data['exitCode'].to_i != 0)
            return true if pod.restart_policy == 'Never'
            return true if not failed and pod.restart_policy == 'OnFailure'
          end

          return false
        end
      end # Container
    end # Kubernetes::Manager
  end # Kube
end # Dapp
