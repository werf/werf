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
          raise ::Dapp::Error::Config, code: :app_name_required if _app.any? { |app| app._name.nil? } && _app.size > 1
          _app.each do |app|
            raise ::Dapp::Error::Config, code: :app_dimg_not_defined, data: { app: app._name } if app._dimg.nil? && !_dimg.map(&:_name).compact.empty?
            unless _dimg.map(&:_name).include?(app._dimg)
              raise ::Dapp::Error::Config, code: :app_dimg_not_found, data: { app: app._name, dimg: app._dimg }
            end
          end

          [:deployment, :app].each do |directive|
            Array(public_send(:"_#{directive}")).each do |obj|
              [:bootstrap, :before_apply_job].each do |job|
                next if (job_config = obj.public_send("_#{job}")).empty?
                job_dimg = job_config._dimg || obj._dimg
                if job_dimg.nil? && !_dimg.map(&:_name).compact.empty?
                  raise ::Dapp::Error::Config, code: :"#{directive}_#{job}_dimg_not_defined"
                end
                unless _dimg.map(&:_name).include?(job_dimg)
                  raise ::Dapp::Error::Config, code: :"#{directive}_#{job}_dimg_not_found", data: { dimg: job_dimg }
                end
              end
            end
          end
        end
      end
    end
  end
end

::Dapp::Config::Config.send(:include, ::Dapp::Deployment::Config::Config)
