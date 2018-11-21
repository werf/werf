module Dapp
  module Kube
    module Dapp
      module Command
        module Ruby2Go
          def ruby2go_deploy_command(command:, command_options: nil, **options)
            (options[:options] ||= {}).merge!(project_dir: path.to_s, raw_command_options: command_options)
            ruby2go_deploy(command: command, **options).tap do |res|
              raise ::Dapp::Error::Command, code: :ruby2go_deploy_command_failed, data: { command: command, message: res["error"] } unless res["error"].nil?
              break res['data']
            end
          end
        end
      end
    end
  end
end
