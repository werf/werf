module Dapp
  module Build
    module Stage
      # BeforeSetupArtifact
      class BeforeSetupArtifact < ArtifactDefault
        def initialize(application, next_stage)
          @prev_stage = BeforeSetup.new(application, self)
          super
        end
      end # BeforeSetupArtifact
    end # Stage
  end # Build
end # Dapp
