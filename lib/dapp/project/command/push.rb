module Dapp
  # Project
  class Project
    # Command
    module Command
      # Push
      module Push
        def push(repo)
          if with_stages?
            log_step('stages')
            with_log_indent { stages_push(repo) }
          end
          build_configs.each do |config|
            log_step(config._name)
            with_log_indent do
              Application.new(config: config, project: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
                app.export!(repo, format: '%{repo}:%{application_name}-%{tag}')
              end
            end
          end
        end

        def with_stages?
          !!cli_options[:with_stages]
        end
      end
    end
  end # Project
end # Dapp
