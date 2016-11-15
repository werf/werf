module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # CleanupLocal
        module CleanupLocal
          def stages_cleanup_local(repo)
            lock_repo(repo, readonly: true) do
              registry = registry(repo)
              repo_dimgs = repo_dimgs_images(registry)
              proper_cache if proper_cache_version?
              cleanup_project(repo_dimgs)
            end
          end

          protected

          def cleanup_project(repo_dimgs)
            lock("#{name}.images") do
              log_step_with_indent(name) do
                project_containers_flush
                project_dangling_images_flush
                dimgs, stages = project_images_hash.partition { |_, image_id| repo_dimgs.values.include?(image_id) }
                dimgs = dimgs.to_h
                stages = stages.to_h
                dimgs.each { |_, aiid| clear_stages(aiid, stages) }
                remove_images(stages.keys)
              end
            end
          end

          def repo_dimgs_images(registry)
            repo_images(registry).first
          end

          def proper_cache
            log_proper_cache do
              lock("#{name}.images") do
                log_step_with_indent(name) do
                  project_containers_flush
                  actual_cache_images = actual_cache_images
                  remove_images(project_images.lines.select { |id| !actual_cache_images.lines.include?(id) }.map(&:strip))
                end
              end
            end
          end

          def actual_cache_images
            shellout!([
              'docker images',
              '--format="{{.Repository}}:{{.Tag}}"',
              %(-f "label=dapp-cache-version=#{Dapp::BUILD_CACHE_VERSION}"),
              stage_cache
            ].join(' ')).stdout.strip
          end

          def project_images_hash
            shellout!(%(docker images --format "{{.Repository}}:{{.Tag}};{{.ID}}" --no-trunc #{stage_cache})).stdout.lines.map do |line|
              line.strip.split(';')
            end.to_h
          end

          def clear_stages(image_id, stages)
            if image_exist?(image_id)
              image_dapp_artifacts_label(image_id).each { |aiid| clear_stages(aiid, stages) }
              iid = image_id
              loop do
                stages.delete_if { |_, siid| siid == iid }
                break if (iid = image_parent(iid)).empty?
              end
            else
              stages.delete_if { |_, siid| siid == image_id }
            end
          end

          def image_exist?(image_id)
            shellout!(%(docker inspect #{image_id}))
            true
          rescue Error::Shellout
            false
          end

          def image_dapp_artifacts_label(image_id)
            select_dapp_artifacts_ids(Image::Docker.image_config_option(image_id: image_id, option: 'labels'))
          end

          def image_parent(image_id)
            shellout!(%(docker inspect -f {{.Parent}} #{image_id})).stdout.strip
          end
        end
      end
    end
  end # Project
end # Dapp
