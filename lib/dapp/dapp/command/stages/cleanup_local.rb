module Dapp
  # Dapp
  class Dapp
    # Command
    module Command
      module Stages
        # CleanupLocal
        module CleanupLocal
          def stages_cleanup_local(repo)
            lock_repo(repo, readonly: true) do
              raise Error::Dapp, code: :stages_cleanup_required_option unless stages_cleanup_option?

              dapp_containers_flush

              proper_cache                 if proper_cache_version?
              stages_cleanup_by_repo(repo) if proper_repo_cache?
              proper_git_commit            if proper_git_commit?
            end
          end

          protected

          def proper_cache
            log_proper_cache do
              lock("#{name}.images") do
                log_step_with_indent(name) do
                  remove_images(dapp_images_names.select { |image_name| !actual_cache_images.include?(image_name) })
                end
              end
            end
          end

          def stages_cleanup_by_repo(repo)
            registry = registry(repo)
            repo_dimgs = repo_dimgs_images(registry)

            lock("#{name}.images") do
              log_step_with_indent(name) do
                dapp_dangling_images_flush
                dimgs, stages = dapp_images_hash.partition { |_, image_id| repo_dimgs.values.include?(image_id) }
                dimgs = dimgs.to_h
                stages = stages.to_h
                dimgs.each { |_, aiid| except_image_with_parents(aiid, stages) }
                remove_images(stages.keys)
              end
            end
          end

          def repo_dimgs_images(registry)
            repo_dimgs_and_cache(registry).first
          end

          def actual_cache_images
            shellout!([
              'docker images',
              '--format="{{.Repository}}:{{.Tag}}"',
              %(-f "label=dapp-cache-version=#{::Dapp::BUILD_CACHE_VERSION}"),
              stage_cache
            ].join(' ')).stdout.lines.map(&:strip)
          end

          def dapp_images_hash
            shellout!(%(docker images --format "{{.Repository}}:{{.Tag}};{{.ID}}" --no-trunc #{stage_cache})).stdout.lines.map do |line|
              line.strip.split(';')
            end.to_h
          end

          def except_image_with_parents(image_id, stages)
            if image_exist?(image_id)
              image_dapp_artifacts_label(image_id).each { |aiid| except_image_with_parents(aiid, stages) }
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
            select_dapp_artifacts_ids(::Dapp::Dimg::Image::Docker.image_config_option(image_id: image_id, option: 'labels'))
          end

          def image_parent(image_id)
            shellout!(%(docker inspect -f {{.Parent}} #{image_id})).stdout.strip
          end

          def proper_git_commit
            log_proper_git_commit do
              unproper_images_names = []
              dapp_images_detailed.each do |_, attrs|
                attrs['Config']['Labels'].each do |repo_name, commit|
                  next if (repo = dapp_git_repositories[repo_name]).nil?
                  unproper_images_names.concat(image_hierarchy_by_id(attrs['Id'])) unless repo.commit_exists?(commit)
                end
              end
              remove_images(unproper_images_names.uniq)
            end
          end

          def dapp_images_detailed
            @dapp_images_detailed ||= {}.tap do |images|
              dapp_images_names.each do |image_name|
                shellout!(%(docker inspect --format='{{json .}}' #{image_name})).stdout.strip.tap do |output|
                  images[image_name] = output == 'null' ? {} : JSON.parse(output)
                end
              end
            end
          end

          def image_hierarchy_by_id(image_id)
            hierarchy = []
            iids = [image_id]

            loop do
              hierarchy.concat(dapp_images_detailed.map { |name, attrs| name if iids.include?(attrs['Id']) }.compact)
              break if begin
                iids.map! do |iid|
                  dapp_images_detailed.map { |_, attrs| attrs['Id'] if attrs['Parent'] == iid }.compact
                end.flatten!.empty?
              end
            end

            hierarchy
          end
        end
      end
    end
  end # Dapp
end # Dapp
