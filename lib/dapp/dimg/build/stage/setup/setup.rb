module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class Setup < Base
            include Mod::Group
            include Mod::GitArtifactsDependencies

            def initialize(dimg, next_stage)
              @prev_stage = GAPreSetupPatch.new(dimg, self)
              super
            end

            def empty?
              !dimg.builder.setup?
            end

            def context
              [git_artifacts_dependencies, builder_checksum]
            end

            def builder_checksum
              dimg.builder.setup_checksum
            end

            def prepare_image
              super
              dimg.builder.setup(image)
            end
          end # Setup
        end # Setup
      end # Stage
    end # Build
  end # Dimg
end # Dapp
