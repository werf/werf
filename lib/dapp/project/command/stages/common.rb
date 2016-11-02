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
            format = proc do |arr|
              arr.map do |tag|
                if (id = registry.image_id(tag)).nil?
                  log_warning(desc: { code: 'tag_ignored', context: 'warning', data: { tag: tag } })
                else
                  [tag, id]
                end
              end.compact.to_h
            end
            dimgs, stages = registry.tags.partition { |tag| !tag.start_with?('dimgstage') }
            [format.call(dimgs), format.call(stages)]
          end

          def delete_repo_image(registry, image_tag)
            if dry_run?
              log(image_tag)
            else
              registry.image_delete(image_tag)
            end
          end

          def select_dapp_artifacts_ids(hash)
            hash.select { |k, _v| k.start_with?('dapp-artifact') }.values
          end

          def lock_repo(repo, *args, &blk)
            lock("repo.#{hashsum repo}", *args, &blk)
          end
        end
      end
    end
  end # Project
end # Dapp
