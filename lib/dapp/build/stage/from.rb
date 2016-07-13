module Dapp
  module Build
    module Stage
      class From < Base
        def signature
          hashsum [from_image_name, application.config._docker._from_cache_version]
        end

        def build!
          return unless should_be_built?
          if application.show_only
            build_log
          else
            from_image.pull!
            image.build!
          end
        end

        def save_in_cache!
          super
          from_image.rmi! if from_image.pulled? && from_image.exist? && !application.show_only
        end

        private

        def from_image_name
          application.config._docker._from
        end

        def from_image
          DockerImage.new(self.application, name: from_image_name)
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
