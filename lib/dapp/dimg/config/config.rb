module Dapp
  module Dimg
    module Config
      module Config
        def self.included(base)
          base.include(Directive::DimgGroupBase)
        end

        protected

        def dimg_config_validate!
          raise Error::Config, code: :dimg_name_required if _dimg.any? { |dimg| dimg._name.nil? } && _dimg.size > 1
          _dimg.each(&:validate!)
        end
      end
    end
  end
end

::Dapp::Config::Config.send(:include, ::Dapp::Dimg::Config::Config)
