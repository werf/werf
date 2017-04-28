module Dapp
  module Deployment
    module Mod
      module SystemEnvironments
        protected

        def system_environments
          @system_environments ||= begin
            [namespace_system_env_name_prefix, system_env_name_prefix].each_with_object({}) do |prefix, environments|
              ENV.map do |k, v|
                if k.start_with?(prefix)
                  key = k.sub(prefix, '')
                  environments[key] = v unless environments.key?(key)
                end
              end
            end
          end
        end

        def namespace_system_env_name_prefix
          [system_env_name_prefix, namespace.to_s.tr('-', '_').upcase].join('_')
        end

        def system_env_name_prefix
          'DAPP_DEPLOYMENT'
        end
      end
    end
  end
end
