module Dapp
  # Dimg
  class Dimg
    include GitArtifact
    include Path
    include Tags
    include Stages

    include Helper::Sha256

    attr_reader :config
    attr_reader :ignore_git_fetch
    attr_reader :should_be_built
    attr_reader :project

    def initialize(config:, project:, should_be_built: false, ignore_git_fetch: false)
      @config = config
      @project = project

      @ignore_git_fetch = ignore_git_fetch
      @should_be_built = should_be_built

      raise Error::Dimg, code: :dimg_not_built if should_be_built?
    end

    def build!
      with_introspection do
        project.lock("#{project.name}.images", readonly: true) do
          last_stage.build_lock! do
            begin
              last_stage.build!
            ensure
              last_stage.save_in_cache! if last_stage.image.built? || project.dev_mode?
            end
          end
        end
      end
    ensure
      FileUtils.rm_rf(tmp_path)
    end

    def export!(repo, format:)
      project.lock("#{project.name}.images", readonly: true) do
        tags.each do |tag|
          image_name = format % { repo: repo, dimg_name: config._name, tag: tag }
          export_base!(last_stage.image, image_name)
        end
      end
    end

    def export_stages!(repo, format:)
      project.lock("#{project.name}.images", readonly: true) do
        export_images.each do |image|
          image_name = format % { repo: repo, signature: image.name.split(':').last }
          export_base!(image, image_name)
        end
      end
    end

    def export_base!(image, image_name)
      if project.dry_run?
        project.log_state(image_name, state: project.t(code: 'state.push'), styles: { status: :success })
      else
        project.lock("image.#{hashsum image_name}") do
          Dapp::Image::Stage.cache_reset(image_name)
          project.log_process(image_name, process: project.t(code: 'status.process.pushing')) do
            project.with_log_indent do
              image.export!(image_name)
            end
          end
        end
      end
    end

    def import_stages!(repo, format:)
      project.lock("#{project.name}.images", readonly: true) do
        import_images.each do |image|
          begin
            image_name = format % { repo: repo, signature: image.name.split(':').last }
            import_base!(image, image_name)
          rescue Error::Shellout
            next
          end
          break unless project.pull_all_stages?
        end
      end
    end

    def import_base!(image, image_name)
      if project.dry_run?
        project.log_state(image_name, state: project.t(code: 'state.pull'), styles: { status: :success })
      else
        project.lock("image.#{hashsum image_name}") do
          project.log_process(image_name,
                              process: project.t(code: 'status.process.pulling'),
                              status: { failed: project.t(code: 'status.failed.not_pulled') },
                              style: { failed: :secondary }) do
            image.import!(image_name)
          end
        end
      end
    end

    def run(docker_options, command)
      cmd = "docker run #{[docker_options, last_stage.image.name, command].flatten.compact.join(' ')}"
      if project.dry_run?
        project.log(cmd)
      else
        system(cmd) || raise(Error::Dimg, code: :dimg_not_run)
      end
    end

    def stage_image_name(stage_name)
      stages.find { |stage| stage.send(:name) == stage_name }.image.name
    end

    def builder
      @builder ||= Builder.const_get(config._builder.capitalize).new(self)
    end

    def artifact?
      false
    end

    def scratch?
      config._docker._from.nil?
    end

    protected

    def should_be_built?
      should_be_built && begin
        builder.before_dimg_should_be_built_check
        !last_stage.image.tagged?
      end
    end

    def with_introspection
      yield
    rescue Exception::IntrospectImage => e
      data = e.net_status[:data]
      cmd = "docker run -ti --rm --entrypoint #{project.bash_path} #{data[:options]} #{data[:built_id]}"
      system(cmd)
      project.shellout!("docker rmi #{data[:built_id]}") if data[:rmi]
      raise data[:error]
    end
  end # Dimg
end # Dapp
