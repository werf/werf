module Dapp
  module Kube
    module Dapp
      module Command
        module Dismiss
          def kube_dismiss
            lock_helm_release do
              kube_check_helm!
              kube_check_helm_release!
              log_process("Delete release #{kube_release_name}") do
                shellout! "helm delete #{kube_release_name} --purge"
                kubernetes.delete_namespace!(kube_namespace) if options[:with_namespace]
              end
            end
          end

          def kube_check_helm_release!
            pr = shellout("helm list | grep #{kube_release_name}")
            raise Error::Command, code: :helm_release_not_exist, data: { name: kube_release_name } if pr.status == 1 || pr.stdout.empty?
          end
        end
      end
    end
  end
end
