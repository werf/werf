module Dapp
  module Kube
    module Dapp
      module Command
        module Lint
          def kube_chart_name
            chart_spec = yaml_load_file(kube_chart_yaml_path)

            if chart_spec["name"].nil? || chart_spec["name"].empty?
              raise ::Dapp::Error::Command, code: :no_helm_chart_spec_name, data: { name: chart_spec["name"], path: kube_chart_yaml_path, raw_spec: kube_chart_yaml_path.read.strip }
            end

            chart_spec["name"]
          end

          def with_kube_tmp_lint_chart_dir(&blk)
            old_kube_tmp_helm_chart_dir = @kube_tmp_helm_chart_dir
            unless ENV['DAPP_HELM_CHART_DIR']
              @kube_tmp_helm_chart_dir = File.join(Dir.mktmpdir('dapp-helm-lint-', tmp_base_dir), kube_chart_name)
            end

            begin
              with_kube_tmp_chart_dir(&blk)
            ensure
              @kube_tmp_helm_chart_dir = old_kube_tmp_helm_chart_dir
            end
          end

          def kube_lint
            kube_check_helm_chart!
            kube_check_helm_chart_yaml!

            repo = option_repo

            docker_tag = option_tags.first

            with_kube_tmp_lint_chart_dir do
              kube_copy_chart
              kube_generate_helm_chart_tpl
              kube_helm_decode_secrets

              all_values = {}
              [kube_chart_path('values.yaml').expand_path, *kube_values_paths, *kube_tmp_chart_secret_values_paths].each do |values_path|
                all_values = all_values.in_depth_merge(yaml_load_file(values_path)) if File.file?(values_path)
              end

              options[:helm_set_options].each do |opt_spec|
                name, _, value = opt_spec.partition("=")
                keys = name.split(".")

                values = all_values
                keys.each_with_index do |key, ind|
                  if ind == keys.size - 1
                    values[key] = YAML.load(value)
                  else
                    values[key] ||= {}
                    values = values[key]
                  end
                end
              end

              service_values = Helm::Values.service_values_hash(self, repo, kube_namespace, docker_tag, fake: true)
              all_values = all_values.in_depth_merge service_values

              kube_chart_path_for_helm.join("values.yaml").write YAML.dump(all_values)

              shellout! "helm lint --strict #{kube_chart_path_for_helm}"
            end
          end

          def kube_check_helm_chart_yaml!
            raise ::Dapp::Error::Command, code: :chart_yaml_not_found, data: { path: kube_chart_yaml_path } unless kube_chart_yaml_path.exist?
          end

          def kube_chart_yaml_path
            kube_chart_path.join("Chart.yaml")
          end
        end
      end
    end
  end
end
