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

        def build!
          return if image.exist? and !application.show_only
          prev_stage.build! if prev_stage
          build_log
          image.build! unless application.show_only
        end

        def fixate!
          return if image.exist?
          prev_stage.fixate! if prev_stage
          image.tag! unless application.show_only
        end

        def signature
          hashsum [prev_stage.signature, *cache_keys]
        end

        def image
          @image ||= begin
            DockerImage.new(name: image_name, from: from_image).tap do |image|
              yield image if block_given?
            end
          end
        end

        def cache_keys
          [application.conf.cache_version]
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
