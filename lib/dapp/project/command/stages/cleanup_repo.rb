module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        module CleanupRepo
          def stages_cleanup_repo(repo)
            lock(repo.to_s) do
              log_step(repo)
              with_log_indent do
                registry = registry(repo)
                repo_applications, repo_stages = repo_images(registry)
                repo_applications.keys.each { |image_tag| clear_repo_stages(registry, repo_stages, image_tag) }
                repo_stages.keys.each { |image_tag| image_delete(registry, image_tag) }
              end
            end
          end

          protected

          def clear_repo_stages(registry, repo_stages, image_tag)
            repo_image_dapp_artifacts_labels(registry, image_tag).each do |iid|
              itag = image_tag_by_image_id(repo_stages, iid)
              clear_repo_stages(registry, repo_stages, itag) unless itag.nil?
            end

            itag = image_tag
            loop do
              repo_stages.delete_if { |sitag, _| sitag == itag }
              break if itag.nil? || (iid = registry.image_parent_id(itag)).empty?
              itag = image_tag_by_image_id(repo_stages, iid)
            end
          end

          def repo_image_dapp_artifacts_labels(registry, image_tag)
            select_dapp_artifacts_ids(registry.image_labels(image_tag))
          end

          def image_tag_by_image_id(images, image_id)
            images.each { |tag, id| return tag if image_id == id }
            nil
          end
        end
      end
    end
  end # Project
end # Dapp
