module Dapp
  module Dimg
    module Dapp
      module ConfigArtifactGroup
        def artifact_config_by_name(name)
          artifacts_configs[name] || raise(::Dapp::Error::Config, code: :artifact_not_found, data: { name: name })
        end

        def artifact_config(name, artifact_config)
          raise(::Dapp::Error::Config, code: :artifact_already_exists, data: { name: name }) if artifacts_configs.key?(name)
          artifacts_configs[name] = artifact_config
        end

        def artifacts_configs
          @artifacts_configs ||= {}
        end
      end # Dimg
    end # Dapp
  end # Dimg
end # Dapp
