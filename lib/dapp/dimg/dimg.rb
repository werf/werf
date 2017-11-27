module Dapp
  module Dimg
    class Dimg
      include GitArtifact
      include Path
      include Stages

      include Helper::Sha256
      include Helper::Trivia

      attr_reader :config
      attr_reader :ignore_git_fetch
      attr_reader :should_be_built
      attr_reader :dapp

      def initialize(config:, dapp:, should_be_built: false, ignore_git_fetch: false)
        @config = config
        @dapp = dapp

        @ignore_git_fetch = ignore_git_fetch
        @should_be_built = should_be_built

        raise Error::Dimg, code: :dimg_not_built if should_be_built?
      end

      def build!
        with_introspection do
          dapp.lock("#{dapp.name}.images", readonly: true) do
            last_stage.build_lock! do
              begin
                builder.before_build_check
                last_stage.build!
              ensure
                after_stages_build!
              end
            end
          end
        end
      ensure
        cleanup_tmp
      end

      def after_stages_build!
        return unless last_stage.image.built? || dev_mode? || force_save_cache?
        last_stage.save_in_cache!
        artifacts.each { |artifact| artifact.last_stage.save_in_cache! }
      end

      def tag!(tag)
        dapp.lock("#{dapp.name}.images", readonly: true) do
          dimg_name = config._name
          if dapp.dry_run?
            dapp.log_state(dimg_name, state: dapp.t(code: 'state.tag'), styles: { status: :success })
          else
            dapp.log_process(dimg_name, process: dapp.t(code: 'status.process.tagging')) do
              last_stage.image.tag!(tag)
            end
          end
        end
      end

      def export!(repo, format:)
        dapp.lock("#{dapp.name}.images", readonly: true) do
          dapp.option_tags.each do |tag|
            image_name = format(format, repo: repo, dimg_name: config._name, tag: tag)
            export_base!(last_stage.image, image_name)
          end
        end
      end

      def export_stages!(repo, format:)
        dapp.lock("#{dapp.name}.images", readonly: true) do
          export_images.each do |image|
            image_name = format(format, repo: repo, signature: image.name.split(':').last)
            export_base!(image, image_name)
          end
        end
      end

      def export_base!(image, image_name)
        if dapp.dry_run?
          dapp.log_state(image_name, state: dapp.t(code: 'state.push'), styles: { status: :success })
        else
          dapp.lock("image.#{hashsum image_name}") do
            ::Dapp::Dimg::Image::Docker.reset_image_inspect(image_name)

            dapp.log_process(image_name, process: dapp.t(code: 'status.process.pushing')) do
              image.export!(image_name)
            end
          end
        end
      end

      def import_stages!(repo, format:)
        dapp.lock("#{dapp.name}.images", readonly: true) do
          import_images.each do |image|
            begin
              image_name = format(format, repo: repo, signature: image.name.split(':').last)
              import_base!(image, image_name)
            rescue ::Dapp::Error::Shellout => e
              dapp.log_info ::Dapp::Helper::NetStatus.message(e)
              next
            end
            break unless dapp.pull_all_stages?
          end
        end
      end

      def import_base!(image, image_name)
        if dapp.dry_run?
          dapp.log_state(image_name, state: dapp.t(code: 'state.pull'), styles: { status: :success })
        else
          dapp.lock("image.#{hashsum image_name}") do
            dapp.log_process(image_name, process: dapp.t(code: 'status.process.pulling'),
                                         status: { failed: dapp.t(code: 'status.failed.not_pulled') },
                                         style: { failed: :secondary }) do
              image.import!(image_name)
            end
          end
        end
      end

      def run(docker_options, command)
        cmd = "#{dapp.host_docker} run #{[docker_options, last_stage.image.built_id, command].flatten.compact.join(' ')}"
        if dapp.dry_run?
          dapp.log(cmd)
        else
          system(cmd) || raise(Error::Dimg, code: :dimg_not_run)
        end
      end

      def stage_image_name(stage_name)
        stages.find { |stage| stage.name == stage_name }.image.name
      end

      def builder
        @builder ||= Builder.const_get(config._builder.capitalize).new(self)
      end

      def artifacts
        @artifacts ||= artifacts_stages.map { |stage| stage.artifacts.map { |artifact| artifact[:dimg] } }.flatten
      end

      def artifact?
        false
      end

      def scratch?
        config._docker._from.nil?
      end

      def dev_mode?
        dapp.dev_mode?
      end

      def force_save_cache?
        !!dapp.options[:force_save_cache]
      end

      def build_cache_version
        [::Dapp::BUILD_CACHE_VERSION, dev_mode? ? 1 : 0]
      end

      def introspect_image!(image:, options:)
        cmd = "#{dapp.host_docker} run -ti --rm --entrypoint #{dapp.bash_bin} #{options} #{image}"
        system(cmd)
      end

      def cleanup_tmp
        # В tmp-директории могли остаться файлы, владельцами которых мы не являемся.
        # Такие файлы могут попасть туда при экспорте файлов артефакта.
        # Чтобы от них избавиться — запускаем docker-контейнер под root-пользователем
        # и удаляем примонтированную tmp-директорию.
        cmd = "".tap do |cmd|
          cmd << "#{dapp.host_docker} run --rm"
          cmd << " --volume #{dapp.tmp_base_dir}:#{dapp.tmp_base_dir}"
          cmd << " alpine:3.6"
          cmd << " rm -rf #{tmp_path}"
        end
        dapp.shellout! cmd

        artifacts.each(&:cleanup_tmp)
      end

      def stage_should_be_introspected?(name)
        dapp.options[:introspect_stage] == name
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
        introspect_image!(image: data[:built_id], options: data[:options])
        raise data[:error]
      end
    end # Dimg
  end # Dimg
end # Dapp
