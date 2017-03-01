module Dapp
  module Build
    module Stage
      # AfterSetupArtifact
      class AfterSetupArtifact < ArtifactDefault
        def initialize(dimg, next_stage)
          @prev_stage = if dimg.artifact?
                          SetupGroup::Setup.new(dimg, self)
                        else
                          SetupGroup::GAPostSetupPatch.new(dimg, self)
                        end
          super
        end
      end # AfterSetupArtifact
    end # Stage
  end # Build
end # Dapp
