module Dapp
  module Deployment
    module Config
      module Config
        include Directive::Mod::Group

        def deployment(&blk)
          directive_eval(_deployment, &blk)
        end

        def _deployment
          @_deployment ||= Directive::Deployment.new(dapp: dapp).tap { |group| @_group << group }
        end

        protected

        def deployment_config_validate!
          [:bootstrap, :before_apply_job].each do |job|
            next if (directive_config = _deployment.public_send("_#{job}")).empty?
            if directive_config._dimg.nil? && !_dimg.map(&:_name).compact.empty?
              raise Error::Config, code: :"deployment_#{job}_dimg_not_defined"
            end
            unless _dimg.map(&:_name).include?(directive_config._dimg)
              raise Error::Config, code: :"deployment_#{job}_dimg_not_found", data: { dimg: directive_config._dimg }
            end
          end

          raise Error::Config, code: :app_name_required if _app.any? { |app| app._name.nil? } && _app.size > 1
          _app.each do |app|
            raise Error::Config, code: :app_dimg_not_defined, data: { app: app._name } if app._dimg.nil? && !_dimg.map(&:_name).compact.empty?
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
