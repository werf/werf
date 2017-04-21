module Dapp
  module Deployment
    module Mod
      module Namespace
        def environment
          config._environment
            .merge(namespace_config_directive(:_environment) || {})
            .merge(system_environments)
        end

        def secret_environment
          return {} if secret.nil?
          config._secret_environment
            .merge(namespace_config_directive(:_secret_environment) || {})
            .each_with_object({}) { |(k, v), environments| environments[k] = secret.extract(v) }
        end

        def scale
          namespace_config_directive(:_scale) || config._scale || 1
        end

        def namespace_config_directive(name)
          return if namespace_config.nil?
          namespace_config.send(name)
        end

        def namespace_config
          @namespace_config = begin
            error_class = Error.const_get(self.class.name.to_s.split('::').last)
            if config._namespace.include?(namespace)
              config._namespace[namespace]
            elsif namespace.nil? || namespace == 'default'
              if config._namespace.one?
                config._namespace.values.first
              elsif config._namespace.empty?
              else
                if namespace.nil?
                  raise error_class, code: :namespace_not_defined, data: { name: name }
                else
                  raise error_class, code: :namespace_not_found, data: { namespace: namespace, name: name }
                end
              end
            else
              raise error_class, code: :namespace_not_found, data: { namespace: namespace, name: name }
            end
          end
        end
      end
    end
  end
end
