module Dapp
  module Dimg
    module Build
      module Stage
        module Install
          class Install < Base
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPreInstallPatch.new(dimg, self)
              super
            end

            def empty?
              !dimg.builder.install?
            end

            def context
              [git_artifacts_dependencies, builder_checksum]
            end

            def builder_checksum
              dimg.builder.install_checksum
            end

            def prepare_image
              super
              dimg.builder.install(image)
            end
          end # Install
        end # Install
      end # Stage
    end # Build
  end # Dimg
end # Dapp
