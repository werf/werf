module Dapp
  module Dimg
    module Build
      module Stage
        module Install
          class GAPreInstallPatch < GARelatedBase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAArchive.new(dimg, self)
              super
            end

            def related_stage_name
              :install
            end
          end # GAPostInstallPatch
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
