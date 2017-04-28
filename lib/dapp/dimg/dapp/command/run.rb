module Dapp
  module Dimg
    module Dapp
      module Command
        module Run
          def run(docker_options, command)
            one_dimg!
            setup_ssh_agent
            Dimg.new(config: build_configs.first, dapp: self, ignore_git_fetch: true, should_be_built: true).run(docker_options, command)
          end
        end
      end
    end
  end # Dimg
end # Dapp
