module Dapp
  module Build
    module Stage
      # AfterSetupArtifact
      class AfterSetupArtifact < ArtifactDefault
        def initialize(application, next_stage)
          @prev_stage = if application.artifact?
                          SetupGroup::ChefCookbooks.new(application, self)
                        else
                          SetupGroup::GAPostSetupPatch.new(application, self)
                        end
          super
        end
      end # AfterSetupArtifact
    end # Stage
  end # Build
end # Dapp
