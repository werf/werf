module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # Common
        module Common
          protected

          def registry(repo)
            DockerRegistry.new(repo)
          end

          def repo_images(registry)
            format = lambda do |arr|
              arr.map do |tag|
                if (id = registry.image_id(tag)).nil?
                  log_warning(desc: { code: 'tag_ignored', context: 'warning', data: { tag: tag } })
                else
                  [tag, id]
                end
              end.compact.to_h
            end
            applications, stages = registry.tags.partition { |tag| !tag.start_with?('dappstage') }
            [format.call(applications), format.call(stages)]
          end

          def repo_image_delete(registry, image_tag)
            if dry_run?
              log(image_tag)
            else
              registry.image_delete(image_tag)
            end
          end

          def select_dapp_artifacts_ids(hash)
            hash.select { |k, _v| k.start_with?('dapp-artifact') }.values
          end

          def proper_cache
            proper_base do
              build_configs.map(&:_basename).uniq.each do |basename|
                lock("#{basename}.images") do
                  log(basename)
                  with_log_indent do
                    containers_flush(basename)
                    project_cache = project_images(basename)
                    proper_cache_images = proper_cache_images(basename)
                    remove_images(project_cache.lines.select { |id| !proper_cache_images.lines.include?(id) }.map(&:strip))
                  end
                end
              end
            end
          end

          def proper_base
            log_step('proper cache')
            with_log_indent do
              yield
            end
          end

          def proper_cache_images(basename)
            shellout!([
              'docker images',
              '--format="{{.Repository}}:{{.Tag}}"',
              %(-f "label=dapp-cache-version=#{Dapp::BUILD_CACHE_VERSION}"),
              stage_cache(basename)
            ].join(' ')).stdout.strip
          end

          def lock_repo(repo, *args, &blk)
            lock("repo.#{hashsum repo}", *args, &blk)
          end
        end
      end
    end
  end # Project
end # Dapp
