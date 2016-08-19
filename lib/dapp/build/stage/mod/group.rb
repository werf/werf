module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Group
        module Group
          def group_name
            class_to_lowercase(self.class.name.split('::')[-2])
          end

          def name
            :"#{group_name}/#{super}"
          end
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
