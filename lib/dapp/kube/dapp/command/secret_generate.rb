module Dapp
  module Kube
    module Dapp
      module Command
        module SecretGenerate
          def kube_secret_generate
            raise Error::Command, code: :secret_key_not_found if secret.nil?
            print 'Enter secret: '
            unless (data = $stdin.noecho(&:gets)).nil?
              puts secret.generate(data.chomp)
            end
          end
        end
      end
    end
  end
end
