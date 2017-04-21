module Dapp
  module Deployment
    class KubeDeployment < KubeBase
      attr_reader :deployment

      def initialize(deployment)
        @deployment = deployment
      end

      def labels
        {'dapp' => deployment.dapp.name, 'dapp-deployment-version' => 1.to_s}
      end

      def delete_unknown_resources!
        # Удаление объектов, связанных с несуществующими более app'ами
        known_apps_names = deployment.apps.map(&:name)
        deployment.kubernetes.deployment_list(labelSelector: labelSelector)['items']
          .reject do |item|
            known_apps_names.include? item['metadata']['labels']['dapp-app']
          end
          .each do |item|
            deployment.dapp.log "Deleting Deployment '#{item['metadata']['name']}': unknown app '#{item['metadata']['labels']['dapp-app']}'"
            deployment.kubernetes.delete_deployment!(item['metadata']['name'])
          end
        deployment.kubernetes.service_list(labelSelector: labelSelector)['items']
          .reject do |item|
            known_apps_names.include? item['metadata']['labels']['dapp-app']
          end
          .each do |item|
            deployment.dapp.log "Deleting Service '#{item['metadata']['name']}': unknown app '#{item['metadata']['labels']['dapp-app']}'"
            deployment.kubernetes.delete_service!(item['metadata']['name'])
          end

        # Версионирование объектов из кода dapp, чтобы
        # иметь возможность автоматом удалить объекты,
        # созданные по старой логике
        old_versions_selector = "dapp=#{labels['dapp']},dapp-deployment-version notin (#{labels['dapp-deployment-version']})"
        deployment.kubernetes.deployment_list(labelSelector: old_versions_selector)['items']
          .each do |item|
            deployment.dapp.log "Deleting Deployment '#{item['metadata']['name']}': dapp-deployment-version changed from #{item['metadata']['labels']['dapp-deployment-version']} to #{labels['dapp-deployment-version']}"
            deployment.kubernetes.delete_deployment!(item['metadata']['name'])
          end
        deployment.kubernetes.service_list(labelSelector: old_versions_selector)['items']
          .each do |item|
            deployment.dapp.log "Deleting Service '#{item['metadata']['name']}': dapp-deployment-version changed from #{item['metadata']['labels']['dapp-deployment-version']} to #{labels['dapp-deployment-version']}"
            deployment.kubernetes.delete_service!(item['metadata']['name'])
          end
      end

      def pod_exist?(name)
        deployment.kubernetes.pod?(name, labelSelector: labelSelector)
      end

      def bootstrap_succeeded?(name)
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

      def run_bootstrap!(spec, name)
        current_spec = deployment.kubernetes.create_pod!(spec)

        deployment.dapp.log_process(:pending) do
          loop do
            current_spec = deployment.kubernetes.pod_status(name)
            break if current_spec['status']['phase'] != 'Pending'
            unless current_spec['status']['containerStatuses'].nil?
              current_spec['status']['containerStatuses'].first['state'].each do |_, desc|
                if desc['reason'] == 'ErrImagePull'
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

      protected

      def labelSelector
        labels.map {|key, value| "#{key}=#{value}"}.join(',')
      end
    end
  end
end
