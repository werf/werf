module Dapp
  module Build
    module Stage
      class Base
        include CommonHelper

        attr_accessor :prev_stage, :next_stage
        attr_reader :application

        def initialize(application, next_stage)
          @application = application

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        def should_be_built?
          !image.exist? and !application.show_only
        end

        def build!
          return unless should_be_built?
          prev_stage.build! if prev_stage
          build_log
          image.build! unless application.show_only
        end

        def save_in_cache!
          return if image.exist?
          prev_stage.save_in_cache! if prev_stage
          image.tag! unless application.show_only
        end

        def signature
          hashsum prev_stage.signature
        end

        def image
          @image ||= begin
            DockerImage.new(self.application, name: image_name, from: from_image).tap do |image|
              image.add_volume "#{application.build_path}:#{application.container_build_path}"
              yield image if block_given?
            end
          end
        end

        protected

        def name
          self.class.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join(?_).downcase.to_sym
        end

        def from_image
          prev_stage.image if prev_stage || begin
            raise 'missing from_image'
          end
        end

        def image_name
          "dapp:#{signature}"
        end

        def build_log
          application.log "#{name} #{"[#{ image.exist? ? image_name : '×' }]" if application.show_only}"
          application.with_log_indent(application.show_only) { application.log "#{image.info}" if image.exist? }
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
