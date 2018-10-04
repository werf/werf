module Dapp
  module Dimg
    module DockerRegistry
      class Dimg < Base
        def dimgstages_tags
          ruby2go_docker_registry_command(command: :dimgstage_tags, options: { reference: tag_reference })
        end

        def dimg_tags(dimg_name)
          with_dimg_repository(dimg_name.to_s) do
            ruby2go_docker_registry_command(command: :dimg_tags, options: { reference: tag_reference })
          end
        end

        def nameless_dimg_tags
          ruby2go_docker_registry_command(command: :dimg_tags, options: { reference: tag_reference })
        end

        def image_id(tag, dimg_repository = nil)
          with_dimg_repository(dimg_repository.to_s) { super(tag) }
        end

        def image_parent_id(tag, dimg_repository = nil)
          with_dimg_repository(dimg_repository.to_s) { super(tag) }
        end

        def image_config(tag, dimg_repository = nil)
          with_dimg_repository(dimg_repository.to_s) { super(tag) }
        end

        def gcr_image_delete(tag, dimg_repository = nil)
          with_dimg_repository(dimg_repository.to_s) { super(tag) }
        end

        def image_delete(tag, dimg_repository = nil)
          with_dimg_repository(dimg_repository.to_s) { super(tag) }
        end

        def image_digest(tag, dimg_repository = nil)
          with_dimg_repository(dimg_repository.to_s) { super(tag) }
        end

        protected

        def with_dimg_repository(dimg_repository)
          old_repository = repository
          @repository    = File.join(old_repository, dimg_repository)
          yield
        ensure
          @repository = old_repository
        end
      end
    end # DockerRegistry
  end # Dimg
end # Dapp
