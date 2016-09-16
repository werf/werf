module Dapp
  module Build
    module Stage
      module InstallGroup
        # Install
        class Install < Base
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPreInstallPatch.new(application, self)
            super
          end

          def empty?
            !application.builder.install?
          end

          def context
            [install_dependencies_files_checksum, application.builder.install_checksum]
          end

          def prepare_image
            super
            application.builder.install(image)
          end

          private

          def install_dependencies_files_checksum
            @install_dependencies_files_checksum ||= dependencies_files_checksum(application.config._install_dependencies)
          end
        end # Install
      end
    end # Stage
  end # Build
end # Dapp
