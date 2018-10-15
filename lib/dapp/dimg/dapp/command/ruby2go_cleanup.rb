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

          def ruby2go_cleanup_common_project_options(force: false)
            {
              common_project_options: {
                project_name: name,
                common_options: {
                  dry_run: dry_run?,
                  force: force
                }
              },
            }
          end

          def ruby2go_cleanup_cache_version_options
            {
              cache_version: ::Dapp::BUILD_CACHE_VERSION.to_s
            }
          end

          def ruby2go_cleanup_common_repo_options
            {
              common_repo_options: {
                repository: option_repo,
                dimgs_names: dimgs_names,
                dry_run: dry_run?
              }
            }
          end
        end
      end
    end
  end # Dimg
end # Dapp
