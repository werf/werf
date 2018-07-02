module Dapp
  module Dimg
    module Build
      module Stage
        module Install
          class Install < Instructions
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPreInstallPatch.new(dimg, self)
              super
            end
          end # Install
        end # Install
      end # Stage
    end # Build
  end # Dimg
end # Dapp
