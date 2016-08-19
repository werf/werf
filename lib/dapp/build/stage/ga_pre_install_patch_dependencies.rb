module Dapp
  module Build
    module Stage
      # GAPreInstallPatchDependencies
      class GAPreInstallPatchDependencies < GADependenciesBase
        def initialize(application, next_stage)
          @prev_stage = GAArchive.new(application, self)
          super
        end

        def dependencies
          [install_dependencies_files_checksum, application.builder.install_checksum]
        end

        def empty?
          super || dependencies_empty?
        end

        private

        def install_dependencies_files_checksum
          @install_dependencies_files_checksum ||= dependencies_files_checksum(application.config._install_dependencies)
        end
      end # GAPreInstallPatchDependencies
    end # Stage
  end # Build
end # Dapp
