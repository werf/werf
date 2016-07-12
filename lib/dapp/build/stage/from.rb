module Dapp
  module Build
    module Stage
      class From < Base
        def signature
          hashsum [from_image_name, application.conf.docker._cache_version(:from)]
        end

        def build!
          return if image.exist? and !application.show_only
          from_image.pull_and_set! if !from_image.exist? and !application.show_only
          build_log
          image.build! unless application.show_only
        end

        def fixate!
          super
          from_image.rmi! if from_image.exist? and !application.show_only
        end

        private

        def from_image_name
          application.conf.docker.from.to_s
        end

        def from_image
          DockerImage.new(self.application, name: from_image_name)
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
