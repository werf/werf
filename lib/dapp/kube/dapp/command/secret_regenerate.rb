module Dapp
  module Kube
    module Dapp
      module Command
        module SecretRegenerate
          def kube_secret_regenerate(*secret_values_paths)
            ruby2go_deploy_command(command: :secret_regenerate, options: { old_key: options[:old_secret_key], secret_values_paths: secret_values_paths })
          end
        end
      end
    end
  end
end
