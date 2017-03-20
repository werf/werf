module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class GAPostSetupPatch < GABase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPostSetupPatchDependencies.new(dimg, self)
              super
            end
          end # GAPostSetupPatch
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
