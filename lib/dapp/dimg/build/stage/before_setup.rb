module Dapp
  module Dimg
    module Build
      module Stage
        class BeforeSetup < Base
          def initialize(dimg, next_stage)
            @prev_stage = AfterInstallArtifact.new(dimg, self)
            super
          end

          def empty?
            !dimg.builder.before_setup?
          end

          def context
            [git_artifacts_dependencies, builder_checksum]
          end

          def builder_checksum
            dimg.builder.before_setup_checksum
          end

          def prepare_image
            super
            dimg.builder.before_setup(image)
          end
        end # BeforeSetup
      end # Stage
    end # Build
  end # Dimg
end # Dapp
