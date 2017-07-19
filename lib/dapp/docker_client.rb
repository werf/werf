module Dapp
  class DockerClient
    class Error < ::Dapp::Error::Base
      def initialize(**net_status)
        super(net_status.merge(context: :docker))
      end
    end

    def initialize # TODO
      # @dapp = dapp
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

    def _images(**spec)
      Docker::Image.all(**spec)
    end

    def _image(name)
      Docker::Image.get(name)
    end

    def _image_create(**spec, &blk)
      Docker::Image.create(**spec, &blk)
    end

    def _image_pull(name)
      # ::Dapp.log_secondary_process(dapp.t(code: 'process.image_pull', data: { name: name })) do # TODO
      repo, tag = name.to_s.split(':')
      spec = { fromImage: repo, tag: tag || :latest }
      if verbose?
        _image_create(**spec) do |chunk|
          JSON.parse(chunk).tap do |json|
            puts %w(id status progress).map { |key| json[key] }.compact.join(' ')
          end
        end
      else
        _image_create(**spec)
      end
      # end
    end

    def _image_push(name:, **spec)
      _image(name).push(nil, **spec)
    end

    def _image_tag(name:, **spec)
      _image(name).tag(**spec)
    end

    def _images_export(**spec)
      Docker::Image.save(**spec)
    end

    def _images_import(**spec)
      Docker::Image.import(**spec)
    end

    def _image_remove(name)
      _image(name).remove
    end

    def _image?(name)
      Docker::Image.exist?(name)
    end

    def _container(name)
      Docker::Container.get(name)
    end

    def _container_create(**spec)
      _image_pull(spec[:image]) unless _image?(spec[:image])
      Docker::Container.create(**spec)
    end

    def _container_run(verbose: false, rm: false, **spec)
      c = _container_create(**spec)
      c_id = c.info['id']

      begin
        code = begin
          if verbose? && verbose
            c.tap { c.start.attach { |_, hunk| puts hunk } }
          else
            c.tap { c.start.attach }
          end.wait['StatusCode']
        end

        unless code.zero?
          raise Error, code: :container_command_failed, data: { name: spec[:name] || c_id,
                                                                msg: _container(c_id).logs(stderr: true).strip }
        end
      ensure
        _pretty_container_remove(c_id) if rm
      end
    end

    def _container_commit(name:, **spec)
      _container(name).commit(**spec).info['id']
    end

    def _container_remove(name, **spec)
      _container(name).remove(**spec)
    end

    def _pretty_container_remove(name)
      return unless _container(name)
      _container_remove(name, force: true)
    end

    def _container?(name)
      _container(name)
      true
    rescue Docker::Error::NotFoundError
      false
    end

    def _validate! # TODO
      Docker.validate_version!
    end

    def verbose? # TODO
      true
      # !dapp.log_quiet?
    end
  end
end
