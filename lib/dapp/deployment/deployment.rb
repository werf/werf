module Dapp
  module Deployment
    class Deployment
      include Mod::Namespace
      include Mod::SystemEnvironments
      include Mod::Jobs

      attr_reader :dapp

      def initialize(dapp:)
        @dapp = dapp
      end

      def name(*args)
        [dapp.name, *args].flatten.compact.join('-').gsub('_', '-')
      end

      def kube
        @kube ||= KubeDeployment.new(self)
      end

      def apps
        @apps ||= dapp.apps_configs.map { |app_config| App.new(app_config: app_config, deployment: self) }
      end

      def namespace
        dapp.options[:namespace] || ENV['DAPP_NAMESPACE']
      end

      def kubernetes
        @kubernetes ||= Kubernetes.new(namespace: namespace)
      end

      protected

      def deployment_config
        dapp.config._deployment
      end
      alias config deployment_config

      def secret
        dapp.secret
      end
    end
  end
end
