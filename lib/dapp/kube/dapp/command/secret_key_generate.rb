module Dapp
  module Kube
    module Dapp
      module Command
        module SecretKeyGenerate
          def kube_secret_key_generate
            puts "DAPP_SECRET_KEY=#{Secret.generate_key}"
          end
        end
      end
    end
  end
end
