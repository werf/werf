module Dapp
  module Kube
    module Dapp
      module Command
        module SecretEdit
          def kube_secret_edit(file_path)
            ruby2go_deploy_command(command: :secret_edit, options: kube_secret_edit_command_options(file_path))
          end

          def kube_secret_edit_command_options(file_path)
            {
              tmp_dir: _ruby2go_tmp_dir,
              file_path: file_path,
              values: options[:values]
            }
          end
        end
      end
    end
  end
end
