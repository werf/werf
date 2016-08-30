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
          return yield if application.project.dry_run?

          try_lock = lambda do
            next yield unless should_be_tagged?

            no_lock = false

            application.project.lock("#{application.config._basename}.image.#{image.name}") do
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

        def build!
          return if should_be_skipped?
          prev_stage.build! if prev_stage
          if image_should_be_build?
            prepare_image
            log_image_build(&method(:image_build))
          end
          raise Exception::IntrospectImage, data: { built_id: image.built_id, options: image.send(:prepared_options) } if should_be_introspected?
        end

        def save_in_cache!
          prev_stage.save_in_cache! if prev_stage
          return unless should_be_tagged?
          image.tag!(log_verbose: application.project.log_verbose?,
                     log_time: application.project.log_time?) unless application.project.dry_run?
        end

        def image
          @image ||= begin
            if empty?
              prev_stage.image
            else
              Image::Stage.new(name: image_name, from: from_image, project: application.project)
            end
          end
        end

        def prepare_image
          image.add_service_change_label dapp: application.stage_dapp_label
          image.add_service_change_label 'dapp-cache-version'.to_sym => Dapp::BUILD_CACHE_VERSION
        end

        def image_should_be_build?
          !empty? || application.project.log_verbose?
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

        def image_build
          image.build!(log_verbose: application.project.log_verbose?,
                       log_time: application.project.log_time?,
                       introspect_error: application.project.cli_options[:introspect_error],
                       introspect_before_error: application.project.cli_options[:introspect_before_error])
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
