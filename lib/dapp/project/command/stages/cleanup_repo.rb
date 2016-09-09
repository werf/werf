module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # CleanupRepo
        module CleanupRepo
          def stages_cleanup_repo(repo)
            lock_repo(repo) do
              registry = registry(repo)
              repo_applications, repo_stages = repo_images(registry)
              repo_stages.delete_if { |_, siid| repo_applications.values.include?(siid) } # ignoring stages with apps ids (v2)
              proper_repo_cache(registry, repo_stages) if proper_cache_version?
              log_step_with_indent(repo) do
                repo_applications.each { |image_tag, image_id| clear_repo_stages(registry, repo_stages, image_tag, image_id) }
                repo_stages.keys.each { |image_tag| delete_repo_image(registry, image_tag) }
              end
            end
          end

          protected

          def clear_repo_stages(registry, repo_stages, image_tag, image_id)
            repo_image_dapp_artifacts_labels(registry, image_tag).each do |iid|
              itag = find_image_tag_by_id(repo_stages, iid)
              clear_repo_stages(registry, repo_stages, itag, iid) unless itag.nil?
            end

            itag = image_tag
            iid = image_id
            loop do
              repo_stages.delete_if { |_, siid| siid == iid }
              break if itag.nil? || (iid = registry.image_parent_id(itag)).empty?
              itag = find_image_tag_by_id(repo_stages, iid)
            end
          end

          def repo_image_dapp_artifacts_labels(registry, image_tag)
            select_dapp_artifacts_ids(registry.image_labels(image_tag))
          end

          def repo_image_dapp_cache_version_label(registry, image_tag)
            registry.image_labels(image_tag)['dapp-cache-version']
          end

          def find_image_tag_by_id(images, image_id)
            images.each { |tag, id| return tag if id == image_id }
            nil
          end

          def proper_repo_cache(registry, repo_stages)
            log_proper_cache do
              wrong_cache_images = repo_stages.select do |image_tag, _|
                repo_image_dapp_cache_version_label(registry, image_tag) != Dapp::BUILD_CACHE_VERSION.to_s
              end
              wrong_cache_images.each { |image_tag, _| delete_repo_image(registry, image_tag) }
              repo_stages.delete_if { |image_tag, _| wrong_cache_images.keys.include?(image_tag) }
            end
          end
        end
      end
    end
  end # Project
end # Dapp
