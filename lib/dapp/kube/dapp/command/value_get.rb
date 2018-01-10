module Dapp
  module Kube
    module Dapp
      module Command
        module ValueGet
          def kube_value_get(value_key)
            service_values = Helm::Values.service_values_hash(
              self, option_repo, kube_namespace, option_tags.first,
              without_registry: self.options[:without_registry],
              disable_warnings: true
            )

            res = service_values
            value_key.split(".").each do |value_key_part|
              if res.is_a?(Hash) && res.key?(value_key_part)
                res = res[value_key_part]
              else
                exit(1)
              end
            end

            if res.is_a? Hash
              puts YAML.dump(res)
            else
              puts JSON.dump(res)
            end
          end
        end
      end
    end
  end
end
