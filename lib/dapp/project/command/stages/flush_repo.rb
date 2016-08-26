module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # FlushRepo
        module FlushRepo
          def stages_flush_repo(repo)
            log_step(repo)
            with_log_indent do
              registry = registry(repo)
              repo_applications, repo_stages = repo_images(registry)
              repo_applications.merge(repo_stages).keys.each { |image_tag| repo_image_delete(registry, image_tag) }
            end
          end
        end
      end
    end
  end # Project
end # Dapp
