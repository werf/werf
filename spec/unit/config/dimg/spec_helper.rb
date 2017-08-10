module SpecHelper
  module Config::Dimg
    include SpecHelper::Config

    def dimg_config_by_name(name)
      dimgs_configs.find { |dimg_config| dimg_config._name == name } || raise
    end

    def dimg_config
      dimgs_configs.first
    end

    def dimgs_configs
      config._dimg
    end

    def dimg_config_validate!
      config.send(:dimg_config_validate!)
    end
  end
end
