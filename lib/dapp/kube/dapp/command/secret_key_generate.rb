module Dapp
  module Kube
    module Dapp
      module Command
        module SecretKeyGenerate
          def kube_secret_key_generate
            puts Secret.generate_key
          end
        end
      end
    end
  end
end
