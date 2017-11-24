module Dapp
  module Dimg
    module Build
      module Stage
        class BeforeInstallArtifact < ArtifactDefault
          def initialize(dimg, next_stage)
            @prev_stage = BeforeInstall.new(dimg, self)
            super
          end

          def image_should_be_untagged_condition
            false
          end
        end # BeforeInstallArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
