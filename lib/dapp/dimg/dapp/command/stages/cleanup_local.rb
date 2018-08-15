module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module CleanupLocal
            def stages_cleanup_local
              lock_repo(option_repo, readonly: true) do
                raise ::Dapp::Error::Command, code: :stages_cleanup_required_option unless stages_cleanup_option?

                proper_cache           if proper_cache_version?
                stages_cleanup_by_repo if proper_repo_cache?
                proper_git_commit      if proper_git_commit?

                dapp_containers_flush_by_label("dapp=#{name}")
                dapp_dangling_images_flush_by_label("dapp=#{name}")
              end
            end

            protected

            def proper_cache
              log_proper_cache do
                lock("#{name}.images") do
                  remove_project_images begin
                    dapp_project_dimgstages.select do |image|
                      !actual_cache_project_dimgstages.map { |dimgstage| dimgstage["Id"] }.include?(image["Id"])
                    end
                  end
                end
              end
            end

            def actual_cache_project_dimgstages
              @actual_cache_project_dimgstages ||= begin
                filters = [{ label: "dapp-cache-version=#{::Dapp::BUILD_CACHE_VERSION}", reference: stage_cache }]
                prepare_docker_images(extra_filters: filters)
              end
            end

            def stages_cleanup_by_repo
              log_proper_repo_cache do
                lock("#{name}.images") do
                  registry   = dimg_registry(option_repo)
                  repo_dimgs = repo_detailed_dimgs_images(registry)
                  dimgstages = clone_dapp_project_dimgstages

                  repo_dimgs.each { |repo_dimg| except_image_id_with_parents(repo_dimg[:parent], dimgstages) }

                  # Удаление только образов старше 2ч
                  dimgstages.delete_if do |dimgstage|
                    Time.now.to_i - dimgstage["Created"] < 2 * 60 * 60
                  end unless ENV['DAPP_STAGES_CLEANUP_LOCAL_DISABLED_DATE_POLICY']

                  remove_project_images(dimgstages)
                end
              end
            end

            def clone_dapp_project_dimgstages
              Marshal.load(Marshal.dump(dapp_project_dimgstages))
            end

            def except_image_id_with_parents(image_id, dimgstages)
              return unless (project_image = dapp_project_image_by_id(image_id))
              except_dapp_project_image_with_parents(project_image, dimgstages)
            end

            def except_dapp_project_image_with_parents(image, dimgstages)
              dapp_project_image_artifacts_ids_in_labels(image).each { |aiid| except_image_id_with_parents(aiid, dimgstages) }
              i = image
              loop do
                array_hash_delete_if_by_id(dimgstages, i)
                break if (i = dapp_project_image_parent(i)).nil?
              end
            end

            def dapp_project_image_artifacts_ids_in_labels(image)
              select_dapp_artifacts_ids(image['Labels'])
            end

            def proper_git_commit
              log_proper_git_commit do
                lock("#{name}.images") do
                  unproper_images = []
                  dapp_project_dimgstages.each do |dimgstage|
                    dimgstage["Labels"].each do |repo_name, commit|
                      next if (repo = dapp_git_repositories[repo_name]).nil?
                      unproper_images.concat(dapp_project_image_with_children(dimgstage)) unless repo.commit_exists?(commit)
                    end
                  end
                  remove_project_images(unproper_images)
                end
              end
            end

            def dapp_project_image_with_children(image)
              children = []
              images   = [image]

              loop do
                children.concat(dapp_project_images.select { |project_image| images.include?(project_image) })
                images.map! do |parent_image|
                  dapp_project_images
                    .select { |project_image| dapp_project_image_parent(project_image) == parent_image }
                end
                images.flatten!
                break if images.empty?
              end

              children
            end

            def dapp_project_image_parent(image)
              dapp_project_image_by_id(image['ParentId'])
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
