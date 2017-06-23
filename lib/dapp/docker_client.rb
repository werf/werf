module Dapp
  class DockerClient
    def initialize
      # @dapp = dapp
      validate!
    end

    def images
      Docker::Image.all
    end

    def image(name)
      Docker::Image.get(name)
    end

    def image_tag(name:, **spec)
      image(name).tag(**spec)
    end

    def image_pull(name:, **spec)
      repo, tag = name.split
      Docker::Image.create(fromImage: repo, tag: tag)
    end

    def image_push(name:, **spec)
      image = Docker::Image.get(name)
      image.push(nil, **spec)
    end

    def images_export(**spec)
      Docker::Image.save(**spec)
    end

    def images_import(**spec)
      Docker::Image.import(**spec)
    end

    def image_remove(**spec)
      Docker::Image.remove(**spec)
    end

    def image?(name)
      Docker::Image.exist?(name)
    end

    def container(name)
      Docker::Container.get(name)
    end

    def container_create(**spec)
      image_pull(name: spec[:image]) unless image?(spec[:image])
      Docker::Container.create(**spec)
    end

    def container_run(verbose: false, rm: false, **spec) # TODO error
      stream = verbose ? STDOUT : ::Dapp::Dapp::Shellout::Streaming::Stream.new
      container = Docker::Container.create(**spec)
      container.tap(&:start).attach { |_, chunk| stream << chunk }
      container_remove(name: container.info['id']) if rm
    end

    def container_commit(name:, **spec)
      container(name).commit(**spec).info['id']
    end

    def container_remove(name:)
      container(name).remove
    end

    def container?(name)
      container(name)
      true
    rescue Docker::Error::NotFoundError
      false
    end

    protected

    attr_reader :dapp

    def validate! # TODO
      Docker.validate_version!
    end
  end
end
