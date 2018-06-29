module Dapp
  module Dimg
    module Build
      module Stage
        class BeforeInstall < Instructions
          def initialize(dimg, next_stage)
            @prev_stage = From.new(dimg, self)
            super
          end

          def dependencies
            [builder_checksum]
          end

          def image_should_be_untagged_condition
            false
          end
        end # BeforeInstall
      end # Stage
    end # Build
  end # Dimg
end # Dapp
