module Dapp
  module Build
    module Stage
      # DockerInstructions
      class DockerInstructions < Base
        def initialize(application)
          @prev_stage = GALatestPatch.new(application, self)
          @application = application
        end

        def dependencies
          [change_options]
        end

        def prepare_image
          super
          change_options.each do |k, v|
            image.public_send("add_change_#{k}", v)
          end
        end
      end # DockerInstructions
    end # Stage
  end # Build
end # Dapp
