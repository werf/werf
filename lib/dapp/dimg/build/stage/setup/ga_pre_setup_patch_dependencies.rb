module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class GAPreSetupPatchDependencies < GARelatedDependenciesBase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = BeforeSetupArtifact.new(dimg, self)
              super
            end

            def related_stage_name
              :setup
            end
          end # GAPreSetupPatchDependencies
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
