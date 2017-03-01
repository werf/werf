module Dapp
  module Dimg
    # Dapp
    class Dapp
      # Command
      module Command
        module Stages
          # CleanupRepo
          module CleanupRepo
            def stages_cleanup_repo(repo)
              lock_repo(repo) do
                raise Error::Command, code: :stages_cleanup_required_option unless stages_cleanup_option?

                registry = registry(repo)
                repo_dimgs, repo_stages = repo_dimgs_and_cache(registry)
                repo_stages.delete_if { |_, siid| repo_dimgs.values.include?(siid) } # ignoring stages with dimgs ids (v2)

                proper_repo_cache(registry, repo_stages)               if proper_cache_version?
                repo_stages_cleanup(registry, repo_dimgs, repo_stages) if proper_repo_cache?
                proper_repo_git_commit(registry)                       if proper_git_commit?
              end
            end

            protected

            def proper_repo_cache(registry, repo_stages)
              log_proper_cache do
                wrong_cache_images = repo_stages.select do |image_tag, _|
                  repo_image_dapp_cache_version_label(registry, image_tag) != ::Dapp::BUILD_CACHE_VERSION.to_s
                end
                wrong_cache_images.each { |image_tag, _| delete_repo_image(registry, image_tag) }
                repo_stages.delete_if { |image_tag, _| wrong_cache_images.keys.include?(image_tag) }
              end
            end

            def repo_stages_cleanup(registry, repo_dimgs, repo_stages)
              log_step_with_indent(repo) do
                repo_dimgs.each { |image_tag, image_id| clear_repo_stages(registry, repo_stages, image_tag, image_id) }
                repo_stages.keys.each { |image_tag| delete_repo_image(registry, image_tag) }
              end
            end

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

            def proper_repo_git_commit(registry)
              log_proper_git_commit do
                unproper_images = []
                repo_dapp_dappstage_images_detailed(registry).each do |_, attrs|
                  attrs[:labels].each do |repo_name, commit|
                    next if (repo = dapp_git_repositories[repo_name]).nil?
                    unproper_images.concat(repo_image_tags_hierarchy(registry, attrs[:id])) unless repo.commit_exists?(commit)
                  end
                end
                remove_repo_images(registry, unproper_images.uniq)
              end
            end

            def repo_dapp_dappstage_images_detailed(registry)
              @repo_dapp_images_detailed ||= begin
                registry.tags.map do |tag|
                  next unless tag.start_with?('dimgstage')

                  image_history = registry.image_history(tag)
                  attrs = {
                    id: registry.image_id(tag),
                    parent: image_history['container_config']['Image'],
                    labels: image_history['config']['Labels']
                  }
                  [tag, attrs]
                end.compact
              end
            end

            def repo_image_tags_hierarchy(registry, registry_image_id)
              hierarchy = []
              iids = [registry_image_id]

              loop do
                hierarchy.concat(iids)
                break if begin
                  iids.map! do |iid|
                    repo_dapp_dappstage_images_detailed(registry).map { |_, attrs| attrs[:id] if attrs[:parent] == iid }.compact
                  end.flatten!.empty?
                end
              end

              repo_dapp_dappstage_images_detailed(registry).map { |tag, attrs| tag if hierarchy.include? attrs[:id] }.compact
            end

            def remove_repo_images(registry, tags)
              tags.each do |tag|
                log(tag) if dry_run? || log_verbose?
                registry.image_delete(tag) unless dry_run?
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
