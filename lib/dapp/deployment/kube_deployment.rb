module Dapp
  module Deployment
    class KubeDeployment
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

      protected

      def labelSelector
        labels.map {|key, value| "#{key}=#{value}"}.join(',')
      end
    end
  end
end
