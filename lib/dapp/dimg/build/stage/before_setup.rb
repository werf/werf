module Dapp
  module Dimg
    module Build
      module Stage
        class BeforeSetup < Instructions
          def initialize(dimg, next_stage)
            @prev_stage = AfterInstallArtifact.new(dimg, self)
            super
          end
        end # BeforeSetup
      end # Stage
    end # Build
  end # Dimg
end # Dapp
