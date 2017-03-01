module Dapp
  module Dimg
    module Build
      module Stage
        # AfterSetupArtifact
        class AfterSetupArtifact < ArtifactDefault
          def initialize(dimg, next_stage)
            @prev_stage = if dimg.artifact?
              Setup::Setup.new(dimg, self)
            else
              Setup::GAPostSetupPatch.new(dimg, self)
            end
            super
          end
        end # AfterSetupArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
