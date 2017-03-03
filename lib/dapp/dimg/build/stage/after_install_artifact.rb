module Dapp
  module Dimg
    module Build
      module Stage
        class AfterInstallArtifact < ArtifactDefault
          def initialize(dimg, next_stage)
            @prev_stage = Install::GAPostInstallPatch.new(dimg, self)
            super
          end
        end # AfterInstallArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
