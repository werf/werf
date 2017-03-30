module Dapp
  module Deployment
    module Config
      module Config
        def self.included(base)
          base.include(Directive::GroupBase)
          base.class_eval do
            alias_method :deployment, :group
            undef_method :group
          end
        end

        protected

        def deployment_config_validate!
          raise Error::Config, code: :app_name_required if _app.any? { |app| app._name.nil? } && _app.size > 1
          _app.each do |app|
            unless _dimg.map(&:_name).include?(app._dimg)
              raise Error::Config, code: :app_dimg_not_found, data: { app: app._name, dimg: app._dimg }
            end
          end
        end
      end
    end
  end
end

::Dapp::Config::Config.send(:include, ::Dapp::Deployment::Config::Config)
