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
            log(basename)
            containers_flush(basename)
            apps, stages = project_images(basename).partition { |_, image_id| repo_applications.values.include?(image_id) }
            apps = apps.to_h
            stages = stages.to_h
            apps.each do |_, aiid|
              iid = aiid
              stages.delete_if { |_, siid| siid == iid } until (iid = image_parent(iid)).empty?
            end
            run_command(%(docker rmi #{stages.keys.join(' ')})) unless stages.keys.empty?
          end
        end

        protected

        def registry(repo)
          @registry ||= DockerRegistry.new(repo).tap { |registry| raise Error::Registry, code: :no_such_app unless registry.repo_exist? }
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
      end
    end
  end # Project
end # Dapp
