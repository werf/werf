module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # Push
        module Push
          def stages_push(repo)
            validate_repo_name(repo)
            build_configs.each do |config|
              log_dimg_name_with_indent(config) do
                Dimg.new(config: config, project: self, ignore_git_fetch: true, should_be_built: true).tap do |dimg|
                  dimg.export_stages!(repo, format: '%{repo}:dimgstage-%{signature}')
                end
              end
            end
          end
        end
      end
    end
  end # Project
end # Dapp
