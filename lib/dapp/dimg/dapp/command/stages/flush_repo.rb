module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module FlushRepo
            def stages_flush_repo
              lock_repo(repo = option_repo) do
                log_step_with_indent("#{repo} stages") do
                  registry = dimg_registry(repo)
                  repo_dimgstages_images(registry).each { |repo_image| delete_repo_image(registry, repo_image) }
                end
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
