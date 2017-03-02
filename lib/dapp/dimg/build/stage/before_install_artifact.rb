module Dapp
  module Dimg
    module Build
      module Stage
        # BeforeInstallArtifact
        class BeforeInstallArtifact < ArtifactDefault
          def initialize(dimg, next_stage)
            @prev_stage = BeforeInstall.new(dimg, self)
            super
          end
        end # BeforeInstallArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
