module Dapp
  module Kube
    module Dapp
      module Command
        module SecretKeyGenerate
          def kube_secret_key_generate
            ruby2go_deploy_command(command: :secret_key_generate)
          end
        end
      end
    end
  end
end
