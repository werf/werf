module Dapp
  module Build
    module Stage
      class From < Base
        def signature
          hashsum [from_image_name, application.conf._docker._cache_version(:from)]
        end

        def build!
          return unless should_be_built?
          from_image.pull! if !from_image.exist? and !application.show_only
          build_log
          image.build! unless application.show_only
        end

        def fixate!
          super
          # FIXME remove image only if it was not exist before build!
          from_image.rmi! if from_image.exist? and !application.show_only
        end

        private

        def from_image_name
          application.conf._docker._from.to_s # FIXME config should do this
        end

        def from_image
          DockerImage.new(self.application, name: from_image_name)
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
