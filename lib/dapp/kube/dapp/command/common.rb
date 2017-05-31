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
            @secret ||= Secret.new(ENV['DAPP_SECRET_KEY']) if ENV.key?('DAPP_SECRET_KEY')
          end

          def kubernetes
            @kubernetes ||= Client.new(namespace: kube_namespace)
          end
        end
      end
    end
  end
end
