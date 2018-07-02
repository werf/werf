module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class Setup < Instructions
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPreSetupPatch.new(dimg, self)
              super
            end
          end # Setup
        end # Setup
      end # Stage
    end # Build
  end # Dimg
end # Dapp
