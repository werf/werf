module Dapp
  module Dimg
    module Dapp
      module Command
        module FlushRepo
          def flush_repo
            lock_repo(repo = option_repo) do
              log_step_with_indent(option_repo) do
                registry = dimg_registry(repo)
                repo_dimgs_images(registry).each { |repo_image| delete_repo_image(registry, repo_image) }
                stages_flush_repo if with_stages?
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
