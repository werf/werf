module Dapp
  module Deployment
    class App
      module SystemEnvironments
        protected

        def system_environments
          @system_environments ||= begin
            namespace_envs = envs(namespace_system_env_name_prefix)
            base_envs      = envs(system_env_name_prefix)
            base_envs.merge(namespace_envs)
          end
        end

        def envs(prefix)
          ENV
            .map { |k, v| [k.sub(prefix, ''), v] if k.start_with?(prefix) }
            .compact
            .to_h
        end

        def namespace_system_env_name_prefix
          [system_env_name_prefix, namespace.to_s.gsub('-', '_').upcase].join('_')
        end

        def system_env_name_prefix
          'DAPP_DEPLOYMENT'
        end
      end
    end
  end
end
