module Dapp
  module Kube
    module Dapp
      module Command
        module Common
          def kube_check_helm!
            raise Error::Command, code: :helm_not_found if shellout('which helm').exitstatus == 1
          end

          def kube_release_name
            "#{name}-#{kube_namespace}"
          end

          def kube_namespace
            options[:namespace].tr('_', '-')
          end

          def secret
            @secret ||= begin
              secret_key = ENV['DAPP_SECRET_KEY']
              secret_key ||= path('.dapp_secret_key').read.chomp if path('.dapp_secret_key').file?
              Secret.new(secret_key) if secret_key
            end
          end

          def kubernetes
            @kubernetes ||= Client.new(namespace: kube_namespace)
          end
        end
      end
    end
  end
end
