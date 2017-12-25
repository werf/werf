module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module CleanupRepo
            def stages_cleanup_repo
              lock_repo(repo = option_repo) do
                raise ::Dapp::Error::Command, code: :stages_cleanup_required_option unless stages_cleanup_option?

                log_step_with_indent("#{repo} stages") do
                  registry        = dimg_registry(repo)
                  repo_dimgs      = repo_dimgs_images(registry)
                  repo_dimgstages = repo_dimgstages_images(registry)

                  repo_dimgstages.delete_if { |dimgstage| repo_dimgs.any? { |dimg| dimgstage[:id] == dimg[:id] } } # ignoring stages with dimgs ids (v2)

                  proper_repo_cache(registry, repo_dimgstages)                   if proper_cache_version?
                  repo_dimgstages_cleanup(registry, repo_dimgs, repo_dimgstages) if proper_repo_cache?
                  proper_repo_git_commit(registry)                               if proper_git_commit?
                end
              end
            end

            protected

            def proper_repo_cache(registry, repo_dimgstages)
              log_proper_cache do
                repo_dimgstages
                  .select { |dimgstage| repo_image_dapp_cache_version_label(registry, dimgstage) != ::Dapp::BUILD_CACHE_VERSION.to_s }
                  .each { |dimgstage| delete_repo_image(registry, dimgstage); repo_dimgstages.delete_at(repo_dimgstages.index(dimgstage)) }
              end
            end

            def repo_dimgstages_cleanup(registry, repo_dimgs, repo_dimgstages)
              log_proper_repo_cache do
                repo_dimgs.each { |dimg| except_repo_image_with_parents(registry, dimg, repo_dimgstages) }
                repo_dimgstages.each { |dimgstage| delete_repo_image(registry, dimgstage) }
              end
            end

            def except_repo_image_with_parents(registry, repo_image, repo_dimgstages)
              repo_image_dapp_artifacts_labels(registry, repo_image).each do |aiid|
                unless (repo_artifact_image = repo_image_by_id(aiid, repo_dimgstages)).nil?
                  except_repo_image_with_parents(registry, repo_artifact_image, repo_dimgstages)
                end
              end

              ri = repo_image
              loop do
                repo_dimgstages.delete_if { |dimgstage| dimgstage == ri }
                ri_parent_id = registry.image_parent_id(ri[:tag], ri[:dimg])
                break if ri_parent_id.empty? || (ri = repo_image_by_id(ri_parent_id, repo_dimgstages)).nil?
              end
            end

            def repo_image_dapp_artifacts_labels(registry, repo_image)
              select_dapp_artifacts_ids(registry.image_labels(repo_image[:tag], repo_image[:dimg]))
            end

            def repo_image_dapp_cache_version_label(registry, repo_image)
              registry.image_labels(repo_image[:tag], repo_image[:dimg])['dapp-cache-version']
            end

            def repo_image_by_id(repo_image_id, repo_images)
              repo_images.find { |repo_image| repo_image[:id] == repo_image_id }
            end

            def proper_repo_git_commit(registry)
              log_proper_git_commit do
                unproper_dimgstages = []
                repo_detailed_dimgstage_images(registry).each do |dimgstage|
                  dimgstage[:labels].each do |repo_name, commit|
                    next if (repo = dapp_git_repositories[repo_name]).nil?
                    unproper_dimgstages.concat(repo_detailed_image_with_children(registry, dimgstage)) unless repo.commit_exists?(commit)
                  end
                end
                unproper_dimgstages.uniq.each { |dimgstage| delete_repo_image(registry, dimgstage) }
              end
            end

            def repo_detailed_dimgstage_images(registry)
              @repo_dapp_dimgstage_images_detailed ||= begin
                repo_dimgstages_images(registry).each do |dimgstage|
                  image_history = registry.image_history(dimgstage[:tag], nil)
                  dimgstage[:parent] = image_history['container_config']['Image']
                  dimgstage[:labels] = image_history['config']['Labels']
                end
              end
            end

            def repo_detailed_image_with_children(registry, image)
              children        = []
              detailed_images = [image]

              loop do
                children.concat(detailed_images)
                detailed_images.map! do |repo_image|
                  repo_detailed_dimgstage_images(registry)
                    .select { |dimgstage| dimgstage[:parent] == repo_image[:id] }
                end
                detailed_images.flatten!
                break if detailed_images.empty?
              end

              children
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
