module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        module CleanupLocal
          def stages_cleanup_local(repo)
            registry = registry(repo)
            repo_applications = repo_applications_images(registry)
            build_configs.map(&:_basename).uniq.each do |basename|
              log(basename)
              with_log_indent do
                containers_flush(basename)
                apps, stages = project_images(basename).partition { |_, image_id| repo_applications.values.include?(image_id) }
                apps, stages = apps.to_h, stages.to_h
                apps.each { |_, aiid| clear_stages(aiid, stages) }
                run_command(%(docker rmi #{stages.keys.join(' ')})) unless stages.keys.empty?
              end
            end
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
              stages.delete_if { |_, siid| siid == iid }
            end
          end

          protected

          def repo_applications_images(registry)
            repo_images(registry).first
          end

          def project_images(basename)
            shellout!(%(docker images --format "{{.Repository}}:{{.Tag}};{{.ID}}" --no-trunc #{stage_cache(basename)})).stdout.lines.map do |line|
              line.strip.split(';')
            end.to_h
          end

          def image_parent(image_id)
            shellout!(%(docker inspect -f {{.Parent}} #{image_id})).stdout.strip
          end

          def image_dapp_artifacts_label(image_id)
            select_dapp_artifacts_ids(Image::Docker.image_config_option(image_id: image_id, option: 'labels'))
          end

          def image_exist?(image_id)
            shellout!(%(docker inspect #{image_id}))
            true
          rescue Error::Shellout
            false
          end
        end
      end
    end
  end # Project
end # Dapp
