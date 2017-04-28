module Dapp
  module Deployment
    class App
      module Namespace
        def environment
          (namespace_config_directive(:_environment) || {})
            .merge(app_config._environment)
            .merge(system_environments)
        end

        def secret_environment
          (namespace_config_directive(:_secret_environment) || {})
            .merge(app_config._secret_environment)
        end

        def scale
          app_config._scale || namespace_config_directive(:_scale) || 1
        end

        def namespace_config_directive(name)
          return if namespace_config.nil?
          namespace_config.send(name)
        end

        def namespace_config
          @namespace_config = begin
            if app_config._namespace.include?(namespace)
              app_config._namespace[namespace]
            elsif namespace.nil? || namespace == 'default'
              if app_config._namespace.one?
                app_config._namespace.first.to_h
              elsif app_config._namespace.empty?
              else
                if namespace.nil?
                  raise Error::App, code: :namespace_not_defined, data: { app_name: name }
                else
                  raise Error::App, code: :namespace_not_found, data: { namespace: namespace, app_name: name }
                end
              end
            else
              raise Error::App, code: :namespace_not_found, data: { namespace: namespace, app_name: name }
            end
          end
        end

        def namespace
          deployment.namespace
        end
      end
    end
  end
end
