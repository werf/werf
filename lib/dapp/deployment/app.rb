module Dapp
  module Deployment
    class App
      include Namespace
      include SystemEnvironments

      attr_reader :app_config
      attr_reader :deployment

      def initialize(app_config:, deployment:)
        @app_config = app_config
        @deployment = deployment
      end

      def dimg
        deployment.dimgs.find { |dimg| dimg.config._name == app_config._dimg }
      end

      [:name, :expose, :bootstrap, :migrate, :run].each do |directive|
        define_method directive do
          app_config.public_send("_#{directive}")
        end
      end

      def deployments
        {}
      end

      def services
        {}
      end
    end
  end
end
