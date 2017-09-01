module Dapp
  module Kube
    module Dapp
      module Command
        module SecretRegenerate
          def kube_secret_regenerate(*secret_values_paths)
            regenerated_data = {}

            secret_values_paths << kube_chart_path('secret-values.yaml') if kube_chart_path('secret-values.yaml').file?
            secret_values_paths.each do |file_path|
              kube_secret_file_validate!(file_path)
              regenerated_data[file_path] = kube_regenerate_secret_values(file_path)
            end

            Dir.glob(kube_chart_secret_path('**/*'), File::FNM_DOTMATCH).each do |file_path|
              next if File.directory?(file_path)
              kube_secret_file_validate!(file_path)
              regenerated_data[file_path] = kube_regenerate_secret_file(file_path)
            end

            regenerated_data.each { |file_path, data| IO.binwrite(file_path, "#{data}\n") }
          end

          def kube_regenerate_secret_file(file_path)
            secret.generate(old_secret.extract(IO.binread(file_path).chomp("\n")))
          end

          def kube_regenerate_secret_values(file_path)
            kube_helm_encode_json(secret, kube_helm_decode_json(old_secret, yaml_load_file(file_path))).to_yaml
          end

          def old_secret
            @old_secret ||= Secret.new(options[:old_secret_key])
          end
        end
      end
    end
  end
end
