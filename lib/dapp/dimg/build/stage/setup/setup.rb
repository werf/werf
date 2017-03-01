module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          # Setup
          class Setup < Base
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPreSetupPatch.new(dimg, self)
              super
            end

            def empty?
              !dimg.builder.setup?
            end

            def context
              [setup_dependencies_files_checksum, builder_checksum]
            end

            def builder_checksum
              dimg.builder.setup_checksum
            end

            def prepare_image
              super
              dimg.builder.setup(image)
            end

            private

            def setup_dependencies_files_checksum
              dependencies_files_checksum(dimg.config._setup_dependencies)
            end
          end # Setup
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
