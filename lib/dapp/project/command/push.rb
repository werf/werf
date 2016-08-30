module Dapp
  # Project
  class Project
    # Command
    module Command
      # Push
      module Push
        def push(repo)
          log_step_with_indent(:stages) { stages_push(repo) } if with_stages?
          build_configs.each do |config|
            log_step_with_indent(config._name) do
              Application.new(config: config, project: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
                app.export!(repo, format: '%{repo}:%{application_name}-%{tag}')
              end
            end
          end
        end

        protected

        def with_stages?
          !!cli_options[:with_stages]
        end
      end
    end
  end # Project
end # Dapp
