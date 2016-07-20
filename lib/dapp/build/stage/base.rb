module Dapp
  module Build
    module Stage
      # Base of all stages
      class Base
        include Helper::Log
        include Helper::Sha256
        include Helper::Trivia

        attr_accessor :prev_stage, :next_stage
        attr_reader :application

        def initialize(application, next_stage)
          @application = application

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        def build!
          return if image.tagged? && !application.log_verbose
          prev_stage.build! if prev_stage
          begin
            if image.tagged?
              application.log_state(name, 'USING CACHE')
            elsif application.dry_run
              application.log_state(name, 'BUILD', styles: { status: :success })
            else
              application.log_process(name, process: 'BUILDING') do
                image_build! if should_be_built?
              end
            end
          ensure
            log_build if application.log?
          end
        end

        def save_in_cache!
          return if image.tagged?
          prev_stage.save_in_cache! if prev_stage
          image.tag!                unless application.dry_run
        end

        def signature
          hashsum prev_stage.signature
        end

        def image
          @image ||= begin
            StageImage.new(name: image_name, from: from_image).tap do |image|
              image.add_volume "#{application.build_path}:#{application.container_build_path}"
              yield image if block_given?
            end
          end
        end

        protected

        def name
          self.class.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join('_').downcase.to_sym
        end

        def should_be_built?
          !image.tagged? && !application.dry_run
        end

        def image_build!
          image.build!(application.log_verbose)
        end

        def from_image
          prev_stage.image if prev_stage || begin
            raise Error::Build, code: :from_image_required
          end
        end

        def image_name
          "dapp:#{signature}"
        end

        def image_info
          date, bytesize = image.info
          _date, from_bytesize = from_image.info
          [date, (from_bytesize.to_i - bytesize.to_i).abs]
        end

        def format_image_info
          date, bytesize = image_info
          ["date: #{Time.parse(date).localtime}", "size: #{to_mb(bytesize.to_i)} MB"].join("\n")
        end

        # rubocop:disable Metrics/AbcSize
        def log_build
          application.with_log_indent do
            application.log_info "signature: #{image_name}"
            application.log_info format_image_info if image.tagged?
            unless (bash_commands = image.send(:bash_commands)).empty?
              application.log_info 'commands:'
              application.with_log_indent { application.log_info bash_commands.join("\n") }
            end
          end if application.log_verbose
        end
        # rubocop:enable Metrics/AbcSize
      end # Base
    end # Stage
  end # Build
end # Dapp
