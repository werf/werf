module Dapp
  # Project
  class Project
    # Command
    module Command
      # StagesCleanup
      module StagesCleanup
        def stages_cleanup(repo)
          repo_applications = repo_applications(repo)
          build_configs.map(&:_basename).uniq.each do |basename|
            lock("#{basename}.images") do
              log(basename)
              containers_flush(basename)
              apps, stages = project_images(basename).partition { |_, image_id| repo_applications.values.include?(image_id) }
              apps, stages = apps.to_h, stages.to_h
              apps.each { |_, aiid| stages = clear_stages(aiid, stages) }
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
            stages.delete_if { |_, siid| siid == image_id }
          end
          stages
        end

        protected

        def registry(repo)
          @registry ||= DockerRegistry.new(repo)
        end

        def repo_applications(repo)
          @repo_apps ||= begin
            registry = registry(repo)
            registry.tags.select { |tag| !tag.start_with?('dappstage') }.map { |tag| [tag, registry.image_id_by_tag(tag)] }.to_h
          end
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
          Image::Docker.image_config_option(image_id: image_id, option: 'labels').select { |k, _v| k.start_with?('dapp-artifact') }.values
        end

        def image_exist?(image_id)
          shellout!(%(docker inspect #{image_id}))
          true
        rescue Error::Shellout
          false
        end
      end
    end
  end # Project
end # Dapp
