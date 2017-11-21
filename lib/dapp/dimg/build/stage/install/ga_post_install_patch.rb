module Dapp
  module Dimg
    module Build
      module Stage
        module Install
          class GAPostInstallPatch < GARelatedBase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = Install.new(dimg, self)
              super
            end

            def related_stage_name
              :before_setup
            end
          end # GAPostInstallPatch
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
