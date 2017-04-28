module Dapp
  module Deployment
    class KubeApp
      attr_reader :app

      def initialize(app)
        @app = app
      end

      def existing_deployments_names
        # TODO: достать из app.deployment.kubernetes.deployment_list
        raise
      end

      def deployment_exist?(name)
        raise
      end

      def deployment_spec(name)
        # NOTICE: Формат должен совпадать с форматом из App::to_kube_deployments
        # NOTICE: Результат из kubernetes-клиент надо привести к тому формату

        raise
      end

      def update_deployment!(spec)
        raise
      end

      def existing_services_names
        raise
      end

      def service_exist?(name)
        raise
      end

      def service_spec(name)
        # NOTICE: Формат должен совпадать с форматом из App::to_kube_services
        # NOTICE: Результат из kubernetes-клиент надо привести к тому формату

        raise
      end

      def update_service!(spec)
        raise
      end
    end
  end
end
