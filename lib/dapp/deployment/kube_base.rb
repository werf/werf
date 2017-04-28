module Dapp
  module Deployment
    class KubeBase
      def pod_exist?(name)
        deployment.kubernetes.pod?(name, labelSelector: labelSelector)
      end

      def pod_succeeded?(name)
        return false unless pod_exist?(name)
        deployment.kubernetes.pod_status(name)['status']['phase'] == 'Succeeded'
      end

      def delete_pod!(name)
        deployment.kubernetes.delete_pod!(name)
        loop do
          break unless deployment.kubernetes.pod?(name)
          sleep 1
        end
      end

      def run_job!(spec, name)
        current_spec = deployment.kubernetes.create_pod!(spec)

        deployment.dapp.log_process(:pending) do
          loop do
            current_spec = deployment.kubernetes.pod_status(name)
            break if current_spec['status']['phase'] != 'Pending'
            unless current_spec['status']['containerStatuses'].nil?
              current_spec['status']['containerStatuses'].first['state'].each do |_, desc|
                if ['ErrImagePull', 'ImagePullBackOff'].include? desc['reason']
                  raise Error::Deployment,
                        code: :bootstrap_image_not_found,
                        data: { reason: desc['reason'], message: desc['message'] }
                end
              end
            end
            sleep 1
          end
        end

        deployment.dapp.log_process(:running, verbose: true) do
          log_thread = Thread.new do
            begin
              deployment.kubernetes.pod_log(name, follow: true) { |chunk| puts chunk }
            rescue Kubernetes::TimeoutError
              deployment.dapp.log_warning('Pod log: read timeout reached!')
            end
          end

          loop do
            current_spec = deployment.kubernetes.pod_status(name)
            break if current_spec['status']['phase'] != 'Running'
            sleep 1
          end

          sleep 1 # last chance to get a message
          log_thread.kill if log_thread.alive?

          phase = current_spec['status']['phase']
          if phase == 'Failed'
            current_spec['status']['containerStatuses'].first['state'].each do |status, desc|
              next if status == 'running'
              raise Error::Deployment,
                    code: :bootstrap_command_failed,
                    data: { reason: desc['reason'], message: desc['message'], exit_code: desc['exitCode'] }
            end
          end
          raise Error::Deployment,
                code: :bootstrap_failed,
                data: { phase: phase, message: current_spec['status']['message'] } unless phase == 'Succeeded'
        end
      end

      def merge_kube_controller_spec(spec1, spec2)
        spec1.kube_in_depth_merge(spec2).tap do |spec|
          spec['spec']['template']['spec']['containers'] = begin
            containers1 = spec1['spec']['template']['spec']['containers']
            containers2 = spec2['spec']['template']['spec']['containers']
            containers2.map do |container2|
              if (container1 = containers1.find { |c| c['name'] == container2['name'] }).nil?
                container2
              else
                container1.kube_in_depth_merge(container2)
              end
            end
          end
        end
      end

      protected

      def labelSelector
        labels.map {|key, value| "#{key}=#{value}"}.join(',')
      end
    end
  end
end
