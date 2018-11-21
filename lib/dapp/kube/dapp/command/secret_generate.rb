module Dapp
  module Kube
    module Dapp
      module Command
        module SecretGenerate
          def kube_secret_generate(file_path)
            ruby2go_deploy_command(command: :secret_generate, command_options: kube_secret_generate_command_options(file_path))
          end

          def kube_secret_generate_command_options(file_path)
            JSON.dump({
              file_path: file_path,
              output_file_path: options[:output_file_path],
              values: options[:values]
            })
          end
        end
      end
    end
  end
end
