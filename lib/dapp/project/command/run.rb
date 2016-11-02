module Dapp
  # Project
  class Project
    # Command
    module Command
      # Run
      module Run
        def run(docker_options, command)
          raise Error::Project, code: :command_unexpected_dimgs_number unless build_configs.one?
          Dimg.new(config: build_configs.first, project: self, ignore_git_fetch: true, should_be_built: true).run(docker_options, command)
        end
      end
    end
  end # Project
end # Dapp
