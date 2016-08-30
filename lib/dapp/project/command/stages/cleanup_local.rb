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
              repo_applications = repo_applications_images(registry)
              proper_cache if proper_cache_version?
              build_configs.map(&:_basename).uniq.each do |basename|
                cleanup_project(basename, repo_applications)
              end
            end
          end

          protected

          def cleanup_project(basename, repo_applications)
            lock("#{basename}.images") do
              log_step_with_indent(basename) do
                project_containers_flush(basename)
                apps, stages = project_images_hash(basename).partition { |_, image_id| repo_applications.values.include?(image_id) }
                apps = apps.to_h
                stages = stages.to_h
                apps.each { |_, aiid| clear_stages(aiid, stages) }
                run_command(%(docker rmi #{stages.keys.join(' ')})) unless stages.keys.empty?
              end
            end
          end

          def repo_applications_images(registry)
            repo_images(registry).first
          end

          def proper_cache
            log_proper_cache do
              build_configs.map(&:_basename).uniq.each do |basename|
                lock("#{basename}.images") do
                  log_step_with_indent(basename) do
                    project_containers_flush(basename)
                    actual_cache_images = actual_cache_images(basename)
                    remove_images(project_images(basename).lines.select { |id| !actual_cache_images.lines.include?(id) }.map(&:strip))
                  end
                end
              end
            end
          end

          def actual_cache_images(basename)
            shellout!([
              'docker images',
              '--format="{{.Repository}}:{{.Tag}}"',
              %(-f "label=dapp-cache-version=#{Dapp::BUILD_CACHE_VERSION}"),
              stage_cache(basename)
            ].join(' ')).stdout.strip
          end

          def project_images_hash(basename)
            shellout!(%(docker images --format "{{.Repository}}:{{.Tag}};{{.ID}}" --no-trunc #{stage_cache(basename)})).stdout.lines.map do |line|
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
