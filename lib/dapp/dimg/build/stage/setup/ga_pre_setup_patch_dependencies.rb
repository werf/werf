module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class GAPreSetupPatchDependencies < GADependenciesBase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = BeforeSetupArtifact.new(dimg, self)
              super
            end

            def dependencies
              dimg.stage_by_name(:setup).context
            end
          end # GAPreSetupPatchDependencies
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
