module Dapp
  module Dimg
    module Build
      module Stage
        class BeforeSetupArtifact < ArtifactDefault
          def initialize(dimg, next_stage)
            @prev_stage = BeforeSetup.new(dimg, self)
            super
          end
        end # BeforeSetupArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
