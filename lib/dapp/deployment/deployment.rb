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

      def dimgs
        @dimgs ||= config._dimg.map do |dimg_config|
          ::Dapp::Dimg::Dimg.new(config: dimg_config, dapp: dapp, should_be_built: true)
        end
      end

      def namespace
        dapp.options[:namespace] || ENV['DAPP_NAMESPACE']
      end
    end
  end
end
