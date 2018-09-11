module Dapp
  module Dimg
    module DockerRegistry
      class Base
        attr_accessor :dapp
        attr_accessor :repository

        def initialize(dapp, repository)
          self.dapp = dapp
          self.repository = repository
        end

        def image_id(tag)
          ruby2go_docker_registry_command(command: :image_id, options: { reference: tag_reference(tag) })
        end

        def image_parent_id(tag)
          ruby2go_docker_registry_command(command: :image_parent_id, options: { reference: tag_reference(tag) })
        end

        def image_config(tag)
          ruby2go_docker_registry_command(command: :image_config, options: { reference: tag_reference(tag) })
        end

        def image_delete(tag)
          digest = image_digest(tag)
          ruby2go_docker_registry_command(command: :image_delete, options: { reference: digest_reference(digest) })
        end

        def image_digest(tag)
          ruby2go_docker_registry_command(command: :image_digest, options: { reference: tag_reference(tag) })
        end

        protected

        def tag_reference(tag = nil)
          [self.repository.chomp("/"), tag].compact.join(":")
        end

        def digest_reference(digest = nil)
          [self.repository.chomp("/"), digest].compact.join("@")
        end

        def ruby2go_docker_registry_command(command:, **options)
          (options[:options] ||= {}).merge!(host_docker_config_dir: dapp.class.host_docker_config_dir)
          dapp.ruby2go_docker_registry(command: command, **options).tap do |res|
            raise Error::Registry, code: :ruby2go_docker_registry_command_failed_unexpected_error, data: { command: command, message: res["error"] } unless res["error"].nil?
            break res['data']
          end
        end
      end
    end # DockerRegistry
  end # Dimg
end # Dapp
