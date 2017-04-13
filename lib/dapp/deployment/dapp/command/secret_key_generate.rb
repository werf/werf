module Dapp
  module Deployment
    module Dapp
      module Command
        module SecretKeyGenerate
          def deployment_secret_key_generate
            puts Secret.generate_key
          end
        end
      end
    end
  end
end
