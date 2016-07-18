module Dapp
  module Build
    module Stage
      # Base of all stages
      class Base
        include CommonHelper

        attr_accessor :prev_stage, :next_stage
        attr_reader :application

        def initialize(application, next_stage)
          @application = application

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        def build!
          return if image.tagged? && !application.dry_run
          prev_stage.build! if prev_stage
          log_build_time do
            log_build
            image.build!(application.logging?) unless application.dry_run
          end
        end

        def save_in_cache!
          return if image.tagged?
          prev_stage.save_in_cache! if prev_stage
          image.tag! unless application.dry_run
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

        def from_image
          prev_stage.image if prev_stage || begin
            raise 'missing from_image'
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
          ["date: #{Time.parse(date)}", "size: #{to_mb(bytesize.to_i)} MB"].join("\n")
        end

        # rubocop:disable Metrics/AbcSize
        def log_build
          application.log "#{name} #{"[#{image.tagged? ? image_name : '×'}]" if application.dry_run}"
          application.with_log_indent(application.dry_run) { application.log format_image_info if image.tagged? }
          bash_commands = image.send(:bash_commands)
          application.with_log_indent do
            application.log('commands:')
            application.with_log_indent { application.log bash_commands.join("\n") }
          end unless bash_commands.empty?
        end
        # rubocop:enable Metrics/AbcSize

        def log_build_time(&blk)
          time = run_time(&blk)
          application.log("build time: #{time.round(2)}") if application.logging? and !application.dry_run
        end

        def run_time
          start = Time.now
          yield
          Time.now - start
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
