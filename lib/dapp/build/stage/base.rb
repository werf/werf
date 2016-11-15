module Dapp
  module Build
    module Stage
      # Base of all stages
      class Base
        include Helper::Sha256
        include Helper::Trivia
        include Mod::Logging

        attr_accessor :prev_stage, :next_stage
        attr_reader :dimg

        def initialize(dimg, next_stage)
          @dimg = dimg

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        # rubocop:disable Metrics/MethodLength, Metrics/PerceivedComplexity
        def build_lock!
          return yield if dimg.project.dry_run?

          try_lock = proc do
            next yield unless should_be_tagged?

            no_lock = false

            dimg.project.lock("#{dimg.project.name}.image.#{image.name}") do
              image.cache_reset

              if should_be_tagged?
                yield
              else
                no_lock = true
              end
            end

            yield if no_lock
          end

          if prev_stage
            prev_stage.build_lock! { try_lock.call }
          else
            try_lock.call
          end
        end
        # rubocop:enable Metrics/MethodLength, Metrics/PerceivedComplexity

        def build!
          return if should_be_skipped?
          prev_stage.build! if prev_stage
          if image_should_be_build?
            prepare_image unless image.tagged?
            log_image_build(&method(:image_build))
          end
          raise Exception::IntrospectImage, data: { built_id: image.built_id, options: image.send(:prepared_options) } if should_be_introspected?
        end

        def save_in_cache!
          prev_stage.save_in_cache! if prev_stage
          return unless should_be_tagged?
          image.save_in_cache! unless dimg.project.dry_run?
        end

        def image
          @image ||= begin
            if empty?
              prev_stage.image
            else
              Image::Stage.new(name: image_name, from: from_image, project: dimg.project)
            end
          end
        end

        def prepare_image
          return if dimg.project.dry_run?
          image.add_volumes_from dimg.project.base_container
          image_add_volumes
          image.add_service_change_label dapp: dimg.stage_dapp_label
          image.add_service_change_label 'dapp-cache-version'.to_sym => Dapp::BUILD_CACHE_VERSION
          image.add_service_change_label 'dapp-dev-mode'.to_sym => true if dimg.dev_mode?

          if dimg.project.ssh_auth_sock
            image.add_volume "#{dimg.project.ssh_auth_sock}:/tmp/dapp-ssh-agent"
            image.add_env 'SSH_AUTH_SOCK', '/tmp/dapp-ssh-agent'
          end
        end

        def image_add_volumes
          image_add_tmp_volumes
          image_add_build_volumes
        end

        def image_add_tmp_volumes
          image_add_default_volumes(:tmp_dir)
        end

        def image_add_build_volumes
          image_add_default_volumes(:build_dir)
        end

        def image_add_default_volumes(type)
          (dimg.config.public_send("_#{type}_mount").map(&:_to) +
            from_image.labels.select { |l, _| l == "dapp-#{type}-dir" }.map { |_, value| value.split(';') }.flatten).each do |path|
            absolute_path = File.expand_path(File.join('/', path))
            tmp_path = dimg.send(type, 'mount', absolute_path[1..-1]).tap(&:mkpath)
            image.add_volume "#{tmp_path}:#{absolute_path}"
          end
        end

        def image_should_be_build?
          !empty? || dimg.project.log_verbose?
        end

        def empty?
          dependencies_empty?
        end

        def dependencies_empty?
          dependencies.flatten.compact.delete_if { |val| val.respond_to?(:empty?) && val.empty? }.empty?
        end

        def signature
          if empty?
            prev_stage.signature
          else
            args = []
            args << prev_stage.signature unless prev_stage.nil?
            args << dimg.build_cache_version
            args << builder_checksum
            args.concat(dependencies.flatten)

            hashsum args
          end
        end

        def builder_checksum
        end

        def dependencies
          []
        end

        protected

        def image_build
          image.build!
        end

        def should_be_skipped?
          image.tagged? && !dimg.project.log_verbose? && !should_be_introspected?
        end

        def should_be_tagged?
          !(empty? || image.tagged? || should_be_not_present?) && image.built?
        end

        def should_be_not_present?
          return false if next_stage.nil?
          next_stage.image.tagged? || next_stage.should_be_not_present?
        end

        def image_name
          dimg.stage_cache_format % { signature: signature }
        end

        def from_image
          prev_stage.image if prev_stage || begin
            raise Error::Build, code: :from_image_required
          end
        end

        def name
          class_to_lowercase.to_sym
        end

        def dependencies_files_checksum(regs)
          regs.map! { |reg| File.directory?(File.join(dimg.project.path, reg)) ? File.join(reg, '**', '*') : reg }
          unless (files = regs.map { |reg| Dir[File.join(dimg.project.path, reg)].map { |f| File.read(f) if File.file?(f) } }).empty?
            hashsum files
          end
        end

        def change_options
          @change_options ||= begin
            dimg.config._docker._change_options.to_h.delete_if do |_, val|
              val.nil? || (val.respond_to?(:empty?) && val.empty?)
            end
          end
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
