module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module CleanupLocal
            def stages_cleanup_local
              lock_repo(option_repo, readonly: true) do
                raise Error::Command, code: :stages_cleanup_required_option unless stages_cleanup_option?

                dapp_project_containers_flush

                proper_cache           if proper_cache_version?
                stages_cleanup_by_repo if proper_repo_cache?
                proper_git_commit      if proper_git_commit?
              end
            end

            protected

            def proper_cache
              log_proper_cache do
                lock("#{name}.images") do
                  log_step_with_indent(name) do
                    remove_project_images(dapp_project_images_ids.select { |image_id| !actual_cache_project_images_ids.include?(image_id) })
                  end
                end
              end
            end

            def actual_cache_project_images_ids
              @actual_cache_project_images_ids ||= begin
                shellout!(%(#{host_docker} images -f "label=dapp" -f "label=dapp-cache-version=#{::Dapp::BUILD_CACHE_VERSION}" #{stage_cache} -q --no-trunc))
                  .stdout
                  .lines
                  .map(&:strip)
              end
            end

            def stages_cleanup_by_repo
              registry = registry(option_repo)
              repo_dimgs = repo_dimgs_images(registry)

              lock("#{name}.images") do
                log_step_with_indent(name) do
                  dapp_project_dangling_images_flush

                  dimgs, stages = dapp_project_images.partition { |image| repo_dimgs.values.include?(image[:id]) }
                  dimgs.each { |dimg_image| except_dapp_project_image_with_parents(dimg_image[:id], stages) }

                  # Удаление только образов старше 2ч
                  stages.delete_if do |stage_image|
                    Time.now - stage_image[:created_at] < 2*60*60
                  end

                  remove_project_images(stages.map { |image| image[:id]} )
                end
              end
            end

            def except_dapp_project_image_with_parents(image_id, stages)
              if dapp_project_image_exist?(image_id)
                dapp_project_image_artifacts_ids_in_labels(image_id).each { |aiid| except_dapp_project_image_with_parents(aiid, stages) }
                iid = image_id
                loop do
                  stages.delete_if { |stage_image| stage_image[:id] == iid }
                  break if (iid = dapp_project_image_parent_id(iid)).nil?
                end
              else
                stages.delete_if { |stage_image| stage_image[:id] == image_id }
              end
            end

            def dapp_project_image_artifacts_ids_in_labels(image_id)
              select_dapp_artifacts_ids(dapp_project_image_labels(image_id))
            end

            def proper_git_commit
              log_proper_git_commit do
                lock("#{name}.images") do
                  dapp_project_dangling_images_flush

                  unproper_images_ids = []
                  dapp_project_images_ids.each do |image_id|
                    dapp_project_image_labels(image_id).each do |repo_name, commit|
                      next if (repo = dapp_git_repositories[repo_name]).nil?
                      unproper_images_ids.concat(dapp_project_image_hierarchy(image_id)) unless repo.commit_exists?(commit)
                    end
                  end
                  remove_project_images(unproper_images_ids)
                end
              end
            end

            def dapp_project_image_hierarchy(image_id)
              hierarchy = []
              iids = [image_id]

              loop do
                hierarchy.concat(dapp_project_images_ids.select { |image_id| iids.include?(image_id) })
                break if begin
                  iids.map! do |iid|
                    dapp_project_images_ids.select { |image_id| dapp_project_image_parent_id(image_id) == iid }
                  end.flatten!.empty?
                end
              end

              hierarchy
            end

            def dapp_project_image_parent_id(image_id)
              dapp_project_image_inspect(image_id)['Parent']
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
