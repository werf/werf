module Dapp
  module Dimg
    module Dapp
      module Command
        module Ruby2GoCleanup
          def ruby2go_cleanup_command(command, command_options)
            options = {
              command: command,
              command_options: command_options,
              options: { host_docker_config_dir: self.class.host_docker_config_dir }
            }

            ruby2go_cleanup(options).tap do |res|
              raise Error::Build, code: :ruby2go_cleanup_command_failed_unexpected_error, data: { command: command, message: res["error"] } unless res["error"].nil?
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
