module Dapp
  module Deployment
    class Deployment
      attr_reader :config
      attr_reader :dapp

      def initialize(config:, dapp:)
        @config = config
        @dapp   = dapp
      end

      def apps
        @apps ||= config._app.map { |app_config| App.new(app_config: app_config, deployment: self) }
      end

      def namespace
        dapp.options[:namespace] || ENV['DAPP_NAMESPACE']
      end

      def kubernetes
        @kubernetes ||= Kubernetes.new(namespace: namespace)
      end
    end
  end
end
