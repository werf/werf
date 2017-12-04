module Dapp
  module Dimg
    module Dapp
      module Command
        module CleanupRepo
          def cleanup_repo
            lock_repo(repo = option_repo) do
              registry = registry(repo)

              repo_detailed_dimgs_images(registry).select do |image|
                case image[:labels]['dapp-tag-scheme']
                  when 'git_tag', 'git_branch', 'git_commit' then true
                  else false
                end && !deployed_docker_images.include?([image[:dimg], image[:tag]].join(':'))
              end.tap do |dimgs_images|
                cleanup_repo_by_nonexistent_git_tag(registry, dimgs_images)
                cleanup_repo_by_nonexistent_git_branch(registry, dimgs_images)
                cleanup_repo_by_nonexistent_git_commit(registry, dimgs_images)
              end

              begin
                registry.reset_cache
                repo_dimgs      = repo_dimgs_images(registry)
                repo_dimgstages = repo_dimgstages_images(registry)
                repo_dimgstages_cleanup(registry, repo_dimgs, repo_dimgstages)
              end if with_stages?
            end
          end

          def cleanup_repo_by_nonexistent_git_tag(registry, dimgs_images)
            cleanup_repo_by_nonexistent_git_base(dimgs_images, 'git_tag') do |dimg_image|
              delete_repo_image(registry, dimg_image) unless git_local_repo.tags.include?(dimg_image[:tag])
            end
          end

          def cleanup_repo_by_nonexistent_git_branch(registry, dimgs_images)
            cleanup_repo_by_nonexistent_git_base(dimgs_images, 'git_branch') do |dimg_image|
              delete_repo_image(registry, dimg_image) unless git_local_repo.remote_branches.include?(dimg_image[:tag])
            end
          end

          def cleanup_repo_by_nonexistent_git_commit(registry, dimgs_images)
            cleanup_repo_by_nonexistent_git_base(dimgs_images, 'git_commit') do |dimg_image|
              delete_repo_image(registry, dimg_image) unless git_local_repo.commit_exists?(dimg_image[:tag])
            end
          end

          def cleanup_repo_by_nonexistent_git_base(dimgs_images, dapp_tag_scheme)
            log_step_with_indent(:"nonexistent #{dapp_tag_scheme.split('_').join(' ')}") do
              dimgs_images
                .select { |dimg_image| dimg_image[:labels]['dapp-tag-scheme'] == dapp_tag_scheme }
                .each { |dimg_image| yield dimg_image }
            end
          end

          def repo_detailed_dimgs_images(registry)
            repo_dimgs_images(registry).each do |dimg|
              image_history = registry.image_history(dimg[:tag], dimg[:dimg])
              dimg[:parent] = image_history['container_config']['Image']
              dimg[:labels] = image_history['config']['Labels']
            end
          end

          def deployed_docker_images
            []
          end
        end
      end
    end
  end # Dimg
end # Dapp
