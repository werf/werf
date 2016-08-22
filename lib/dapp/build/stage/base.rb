module Dapp
  module Build
    module Stage
      # Base of all stages
      class Base
        include Helper::Sha256
        include Helper::Trivia
        include Mod::Logging

        attr_accessor :prev_stage, :next_stage
        attr_reader :application

        def initialize(application, next_stage)
          @application = application

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        def build_lock!
          return yield if application.dry_run?

          try_lock = lambda do
            next yield unless should_be_tagged?
            application.project.lock("#{application.config._basename}.image.#{image.name}") do
              image.cache_reset
              yield
            end
          end

          if prev_stage
            prev_stage.build_lock! { try_lock.call }
          else
            try_lock.call
          end
        end

        def build!
          return if should_be_skipped?
          prev_stage.build! if prev_stage
          image_build if image_should_be_build?
          raise Exception::IntrospectImage, data: { built_id: image.built_id, options: image.send(:prepared_options) } if should_be_introspected?
        end

        def save_in_cache!
          prev_stage.save_in_cache! if prev_stage
          return unless should_be_tagged?
          image.tag!(log_verbose: application.log_verbose?, log_time: application.log_time?) unless application.dry_run?
        end

        def image
          @image ||= begin
            if empty?
              prev_stage.image
            else
              Image::Stage.new(name: image_name, from: from_image).tap do |image|
                image.add_service_change_label dapp: application.stage_dapp_label
                yield image if block_given?
              end
            end
          end
        end

        def image_should_be_build?
          !empty? || application.log_verbose?
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
            hashsum [prev_stage.signature, *dependencies.flatten]
          end
        end

        def dependencies
          []
        end

        protected

        # rubocop:disable Metrics/AbcSize
        def image_build
          if empty?
            application.log_state(log_name, state: application.t(code: 'state.empty'))
          elsif image.tagged?
            application.log_state(log_name, state: application.t(code: 'state.using_cache'))
          elsif should_be_not_present?
            application.log_state(log_name, state: application.t(code: 'state.not_present'))
          elsif application.dry_run?
            application.log_state(log_name, state: application.t(code: 'state.build'), styles: { status: :success })
          else
            application.log_process(log_name, process: application.t(code: 'status.process.building'), short: should_not_be_detailed?) do
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

        def should_be_tagged?
          !(empty? || image.tagged? || should_be_not_present?)
        end

        def should_be_not_present?
          return false if next_stage.nil?
          next_stage.image.tagged? || next_stage.should_be_not_present?
        end

        def image_name
          application.stage_cache_format % { signature: signature }
        end

        def from_image
          prev_stage.image if prev_stage || begin
            raise Error::Build, code: :from_image_required
          end
        end

        def name
          class_to_lowercase.to_sym
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
