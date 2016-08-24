module Dapp
  # Project
  class Project
    # Command
    module Command
      # Run
      module Run
        def run(docker_options, command)
          raise Error::Project, code: :run_command_unexpected_apps_number unless build_configs.one?
          Application.new(config: build_configs.first, project: self, ignore_git_fetch: true).run(docker_options, command)
        end
      end
    end
  end # Project
end # Dapp
