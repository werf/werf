module Dapp
  # Application
  class Application
    include GitArtifact
    include Path
    include Tags

    include Helper::Sha256

    attr_reader :config
    attr_reader :ignore_git_fetch
    attr_reader :should_be_built
    attr_reader :is_artifact
    attr_reader :project

    def initialize(config:, project:, should_be_built: false, ignore_git_fetch: false, is_artifact: false)
      @config = config
      @project = project

      @tmp_path = Dir.mktmpdir(project.cli_options[:tmp_dir_prefix] || 'dapp-')

      @last_stage = Build::Stage::DockerInstructions.new(self)
      @ignore_git_fetch = ignore_git_fetch
      @should_be_built = should_be_built
      @is_artifact = is_artifact

      raise Error::Application, code: :application_not_built if should_be_built?
    end

    def build!
      with_introspection do
        project.lock("#{config._basename}.images", readonly: true) do
          last_stage.build_lock! do
            last_stage.build!
            last_stage.save_in_cache!
          end
        end
      end
    ensure
      FileUtils.rm_rf(tmp_path)
    end

    def export!(repo, format:)
      project.lock("#{config._basename}.images", readonly: true) do
        tags.each do |tag|
          image_name = format % { repo: repo, application_name: config._name, tag: tag }
          export_base!(last_stage.image, image_name)
        end
      end
    end

    def export_stages!(repo, format:)
      project.lock("#{config._basename}.images", readonly: true) do
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
            image.export!(image_name, log_verbose: project.log_verbose?, log_time: project.log_time?)
          end
        end
      end
    end

    def import_stages!(repo, format:)
      project.lock("#{config._basename}.images", readonly: true) do
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
            image.import!(image_name, log_verbose: project.log_verbose?, log_time: project.log_time?)
          end
        end
      end
    end

    def run(docker_options, command)
      cmd = "docker run #{[docker_options, last_stage.image.name, command].flatten.compact.join(' ')}"
      if project.dry_run?
        project.log(cmd)
      else
        system(cmd) || raise(Error::Application, code: :application_not_run)
      end
    end

    def signature
      last_stage.send(:signature)
    end

    def stage_cache_format
      "#{project.cache_format % { application_name: config._basename }}:%{signature}"
    end

    def stage_dapp_label
      project.stage_dapp_label_format % { application_name: config._basename }
    end

    def artifact(config)
      self.class.new(config: config, project: project, ignore_git_fetch: ignore_git_fetch, is_artifact: true, should_be_built: should_be_built)
    end

    def builder
      @builder ||= Builder.const_get(config._builder.capitalize).new(self)
    end

    def export_images
      images.select(&:tagged?)
    end

    def import_images
      images.select { |image| !image.tagged? }
    end

    def images
      (@images ||= []).tap do |images|
        stage = last_stage
        loop do
          if stage.respond_to?(:images)
            images.concat(stage.images)
          else
            images << stage.image
          end
          break if (stage = stage.prev_stage).nil?
        end
      end.uniq(&:name)
    end

    def should_be_built?
      should_be_built && !is_artifact && begin
        builder.before_application_should_be_built_check
        !last_stage.image.tagged?
      end
    end

    protected

    def with_introspection
      yield
    rescue Exception::IntrospectImage => e
      data = e.net_status[:data]
      cmd = "docker run -ti --rm --entrypoint /bin/bash #{data[:options]} #{data[:built_id]}"
      system(cmd).tap do |res|
        project.shellout!("docker rmi #{data[:built_id]}") if data[:rmi]
        res || raise(Error::Application, code: :application_not_run)
      end
      exit 0
    end

    attr_reader :last_stage
  end # Application
end # Dapp
