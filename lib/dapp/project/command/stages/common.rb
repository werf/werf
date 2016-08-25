module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        module Common
          protected

          def registry(repo)
            DockerRegistry.new(repo)
          end

          def repo_images(registry)
            format = ->(arr) do
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

          def image_delete(registry, image_tag)
            if dry_run?
              log(image_tag)
            else
              registry.image_delete(image_tag)
            end
          end

          def select_dapp_artifacts_ids(hash)
            hash.select { |k, _v| k.start_with?('dapp-artifact') }.values
          end
        end
      end
    end
  end # Project
end # Dapp
