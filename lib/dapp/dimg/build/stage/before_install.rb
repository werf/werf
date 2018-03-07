module Dapp
  module Dimg
    module Build
      module Stage
        class BeforeInstall < Base
          def initialize(dimg, next_stage)
            @prev_stage = From.new(dimg, self)
            super
          end

          def empty?
            !dimg.builder.before_install?
          end

          def dependencies
            [builder_checksum]
          end

          def builder_checksum
            dimg.builder.before_install_checksum
          end

          def prepare_image
            super do
              dimg.builder.before_install(image)
            end
          end

          def image_should_be_untagged_condition
            false
          end
        end # BeforeInstall
      end # Stage
    end # Build
  end # Dimg
end # Dapp
