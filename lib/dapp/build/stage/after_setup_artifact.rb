module Dapp
  module Build
    module Stage
      # AfterSetupArtifact
      class AfterSetupArtifact < ArtifactBase
        def initialize(application, next_stage)
          @prev_stage = SetupGroup::GAPostSetupPatch.new(application, self)
          super
        end
      end # AfterSetupArtifact
    end # Stage
  end # Build
end # Dapp
