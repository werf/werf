module Dapp
  module Dimg
    module Build
      module Stage
        class GAArtifactPatch < GALatestPatch
          def initialize(dimg, next_stage)
            @prev_stage = Setup::Setup.new(dimg, self)
            super
          end

          def empty?
            dimg.git_artifacts.empty? || dependencies_empty?
          end

          def dependencies
            dimg.stage_by_name(:build_artifact).context
          end
        end # GAArtifactPatch
      end # Stage
    end # Build
  end # Dimg
end # Dapp
