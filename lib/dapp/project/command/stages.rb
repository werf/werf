module Dapp
  # Project
  class Project
    # Command
    module Command
      # Stages
      module Stages
        def stages_flush
          build_configs.map(&:_basename).uniq.each do |basename|
            log(basename)
            containers_flush(basename)
            remove_images(%(docker images --format="{{.Repository}}:{{.Tag}}" #{stage_cache(basename)}))
          end
        end

        def stages_cleanup(repo)
          repo_apps = repo_apps(repo)
          build_configs.map(&:_basename).uniq.each do |basename|
            log(basename)
            containers_flush(basename)
            apps, stages = project_images(basename).partition { |_, image_id| repo_apps.values.include?(image_id) }
            apps = apps.to_h
            stages = stages.to_h
            apps.each do |_, aiid|
              iid = aiid
              until (iid = image_parent(iid)).empty?
                stages.delete_if { |_, siid| siid == iid }
              end
            end
            run_command(%(docker rmi #{stages.keys.join(' ')})) unless stages.keys.empty?
          end
        end

        protected

        def repo_apps(repo)
          registry = DockerRegistry.new(repo)
          raise Error::Registry, :no_such_app unless registry.repo_exist?
          registry.repo_apps
        end

        def project_images(basename)
          shellout!(%(docker images --format "{{.Repository}}:{{.Tag}};{{.ID}}" --no-trunc #{stage_cache(basename)})).stdout.lines.map do |line|
            line.strip.split(';')
          end.to_h
        end

        def image_parent(image_id)
          shellout!(%(docker inspect -f {{.Parent}} #{image_id})).stdout.strip
        end
      end
    end
  end # Project
end # Dapp
