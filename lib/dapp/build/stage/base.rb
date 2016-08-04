module Dapp
  module Build
    module Stage
      # Base of all stages
      class Base
        include Helper::Sha256
        include Helper::Trivia
        include Mod::Artifact
        include Mod::Logging

        attr_accessor :prev_stage, :next_stage
        attr_reader :application

        def initialize(application, next_stage)
          @application = application

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        def build!
          return if should_be_skipped?
          prev_stage.build! if prev_stage
          image_build
          raise Exception::IntrospectImage,
                message: application.t(code: 'introspect.stage', data: { name: name }),
                data: { built_id: image.built_id, options: image.send(:prepared_options) } if should_be_introspected?
        end

        def save_in_cache!
          return if image.tagged?
          prev_stage.save_in_cache!                                                          if prev_stage
          image.tag!(log_verbose: application.log_verbose?, log_time: application.log_time?) unless application.dry_run?
        end

        def signature
          hashsum [prev_stage.signature, artifacts_signatures]
        end

        def image
          @image ||= begin
            Image::Stage.new(name: image_name, from: from_image).tap do |image|
              image.add_volume "#{application.tmp_path}:#{application.container_tmp_path}"
              before_artifacts.each { |artifact| apply_artifact(artifact, image) }
              yield image if block_given?
              after_artifacts.each { |artifact| apply_artifact(artifact, image) }
            end
          end
        end

        protected

        def name
          class_to_lowercase.to_sym
        end

        # rubocop:disable Metrics/AbcSize
        def image_build
          if image.tagged?
            application.log_state(name, state: application.t(code: 'state.using_cache'))
          elsif should_be_not_present?
            application.log_state(name, state: application.t(code: 'state.not_present'))
          elsif application.dry_run?
            application.log_state(name, state: application.t(code: 'state.build'), styles: { status: :success })
          else
            application.log_process(name, process: application.t(code: 'status.process.building'), short: should_be_not_detailed?) do
              image_do_build
            end
          end
        ensure
          log_build
        end
        # rubocop:enable Metrics/AbcSize

        def image_do_build
          image.build!(log_verbose: application.log_verbose?,
                       log_time: application.log_time?,
                       introspect_error: application.cli_options[:introspect_error],
                       introspect_before_error: application.cli_options[:introspect_before_error])
        end

        def from_image
          prev_stage.image if prev_stage || begin
            raise Error::Build, code: :from_image_required
          end
        end

        def image_name
          "dapp:#{signature}"
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
