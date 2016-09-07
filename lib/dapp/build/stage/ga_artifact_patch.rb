module Dapp
  module Build
    module Stage
      # GAArtifactPatch
      class GAArtifactPatch < GALatestPatch
        def initialize(application, next_stage)
          @prev_stage = SetupGroup::ChefCookbooks.new(application, self)
          super
        end

        def dependencies
          next_stage.dependencies # BuildArtifact
        end

        def prev_g_a_stage
          super.prev_stage.prev_stage # GAPreSetupPatch
        end
      end # GAArtifactPatch
    end # Stage
  end # Build
end # Dapp
