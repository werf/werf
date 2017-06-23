module Dapp
  class DockerClient
    class Error < ::Dapp::Error::Base
      def initialize(**net_status)
        super(net_status.merge(context: :docker))
      end
    end

    def initialize(dapp)
      @dapp = dapp
      validate!
    end

    def method_missing(m, *args, &block)
      if respond_to?(m)
        send(:"_#{m}", *args, &block)
      else
        super
      end
    rescue ::Docker::Error::DockerError => e
      raise Error, code: :exception, data: { type: e.class.to_s.split('::').last, msg: e.message.strip }
    end

    def respond_to_missing?(m, include_private = false)
      methods.include?(:"_#{m}") || super
    end

    protected

    attr_reader :dapp

    def _images
      Docker::Image.all
    end

    def _image(name)
      Docker::Image.get(name)
    end

    def _image_tag(name:, **spec)
      _image(name).tag(**spec)
    end

    def _image_pull(name:)
      dapp.log_secondary_process(dapp.t(code: 'process.image_pull', data: { name: name })) do
        repo, tag = name.split
        kwargs = { fromImage: repo, tag: tag }
        if verbose?
          Docker::Image.create(**kwargs) do |chunk|
            JSON.parse(chunk).tap do |json|
              msg = []
              msg << json['id']       if json.key?('id')
              msg << json['status']   if json.key?('status')
              msg << json['progress'] if json.key?('progress')
              puts msg.join(' ')
            end
          end
        else
          Docker::Image.create(**kwargs)
        end
      end
    end

    def _image_push(name:, **spec)
      image = Docker::Image.get(name)
      image.push(nil, **spec)
    end

    def _images_export(**spec)
      Docker::Image.save(**spec)
    end

    def _images_import(**spec)
      Docker::Image.import(**spec)
    end

    def _image_remove(**spec)
      Docker::Image.remove(**spec)
    end

    def _image?(name)
      Docker::Image.exist?(name)
    end

    def _container(name)
      Docker::Container.get(name)
    end

    def _container_create(**spec)
      _image_pull(name: spec[:image]) unless _image?(spec[:image])
      Docker::Container.create(**spec)
    end

    def _container_run(verbose: false, rm: false, **spec)
      container = _container_create(**spec)
      if verbose? && verbose
        container.tap(&:start).attach { |_, hunk| puts hunk }
      else
        container.tap(&:start).attach
      end
      _container_remove(name: container.info['id']) if rm
    end

    def _container_commit(name:, **spec)
      _container(name).commit(**spec).info['id']
    end

    def _container_remove(name:)
      _container(name).remove
    end

    def _container?(name)
      _container(name)
      true
    rescue Docker::Error::NotFoundError
      false
    end

    def verbose?
      !dapp.log_quiet?
    end

    def validate! # TODO
      Docker.validate_version!
    end
  end
end
