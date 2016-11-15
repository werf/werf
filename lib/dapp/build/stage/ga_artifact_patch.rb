module Dapp
  module Build
    module Stage
      # GAArtifactPatch
      class GAArtifactPatch < GALatestPatch
        def initialize(dimg, next_stage)
          @prev_stage = SetupGroup::ChefCookbooks.new(dimg, self)
          super
        end

        def dependencies
          next_stage.context # BuildArtifact
        end

        def prev_g_a_stage
          super.prev_stage.prev_stage # GAPreSetupPatch
        end
      end # GAArtifactPatch
    end # Stage
  end # Build
end # Dapp
