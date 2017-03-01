module Dapp
  module Build
    module Stage
      module InstallGroup
        # Install
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
            [install_dependencies_files_checksum, builder_checksum]
          end

          def builder_checksum
            dimg.builder.install_checksum
          end

          def prepare_image
            super
            dimg.builder.install(image)
          end

          private

          def install_dependencies_files_checksum
            dependencies_files_checksum(dimg.config._install_dependencies)
          end
        end # Install
      end
    end # Stage
  end # Build
end # Dapp
