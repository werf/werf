module Dapp
  module Dimg
    module Dapp
      module Command
        module Build
          def build
            command = "build"

            # TODO: move project name logic to golang
            project_name = name.to_s

            # TODO: move project dir logic to golang
            project_dir = path.to_s

            res = ruby2go_build(
              "command" => command,
              "projectName" => project_name,
              "projectDir" => project_dir,
              "rubyCliOptions" => JSON.dump(self.options),
              options: {},
            )

            raise ::Dapp::Error::Command, code: :ruby2go_deploy_command_failed, data: { command: command, message: res["error"] } unless res["error"].nil?
          end

          def build_old
            build_configs.each do |config|
              log_dimg_name_with_indent(config) do
                dimg(config: config).build!
              end
            end
          rescue ::Dapp::Error::Shellout, Error::Default
            build_context_export unless options[:build_context_directory].nil?
            raise
          end
        end
      end
    end
  end # Dimg
end # Dapp
