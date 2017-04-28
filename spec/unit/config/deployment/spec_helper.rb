module SpecHelper
  module Config::Deployment
    include SpecHelper::Config

    def app_config
      apps_configs.first
    end

    def apps_configs
      config._app
    end

    def deployment_config_validate!
      config.send(:deployment_config_validate!)
    end
  end
end
