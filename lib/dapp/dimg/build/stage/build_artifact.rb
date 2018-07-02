module Dapp
  module Dimg
    module Build
      module Stage
        class BuildArtifact < Instructions
          def initialize(dimg)
            @prev_stage = GAArtifactPatch.new(dimg, self)
            @dimg = dimg
          end
        end # BuildArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
