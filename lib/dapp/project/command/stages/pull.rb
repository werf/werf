module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # Pull
        module Pull
          def stages_pull(repo)
            build_configs.each do |config|
              log_step(config._name)
              with_log_indent do
                Application.new(config: config, project: self, ignore_git_fetch: true).tap do |app|
                  app.import_stages!(repo, format: '%{repo}:dappstage-%{signature}')
                end
              end
            end
          end

          def pull_all_stages?
            !!cli_options[:pull_all_stages]
          end
        end
      end
    end
  end # Project
end # Dapp
