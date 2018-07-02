module Dapp
  module Dimg
    module Build
      module Stage
        class Instructions < Base
          def empty?
            !dimg.builder.public_send(:"#{name}?")
          end

          def context
            [git_artifacts_dependencies, builder_checksum]
          end

          def builder_checksum
            dimg.builder.public_send(:"#{name}_checksum")
          end

          def prepare_image
            super do
              dimg.builder.public_send(name, image)
            end
          end
        end # Instructions
      end # Stage
    end # Build
  end # Dimg
end # Dapp
