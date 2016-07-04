module Dapp
  module Build
    module Stage
      class From < Base
        def signature
          hashsum from_image_name # TODO: add hashkey
        end

        def build!
          return if image.exist?
          from_image.pull_and_set! unless from_image.exist?
          application.log self.class.to_s
          image.build!
        end

        def fixate!
          super
          from_image.rmi! if from_image.exist?
        end

        private

        def from_image_name
          application.conf[:from].to_s
        end

        def from_image
          DockerImage.new(name: from_image_name)
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
