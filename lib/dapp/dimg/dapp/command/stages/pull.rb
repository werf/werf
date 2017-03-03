module Dapp
  module Dimg
    class Dapp
      module Command
        module Stages
          module Pull
            def stages_pull(repo)
              validate_repo_name(repo)
              build_configs.each do |config|
                log_dimg_name_with_indent(config) do
                  Dimg.new(config: config, dapp: self, ignore_git_fetch: true).tap do |dimg|
                    dimg.import_stages!(repo, format: '%{repo}:dimgstage-%{signature}')
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
    end
  end # Dimg
end # Dapp
