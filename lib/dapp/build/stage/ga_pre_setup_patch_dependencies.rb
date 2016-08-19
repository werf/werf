module Dapp
  module Build
    module Stage
      # GAPreSetupPatchDependencies
      class GAPreSetupPatchDependencies < GADependenciesBase
        def initialize(application, next_stage)
          @prev_stage = BeforeSetup.new(application, self)
          super
        end

        def dependencies
          [setup_dependencies_files_checksum, application.builder.setup_checksum]
        end

        private

        def setup_dependencies_files_checksum
          @setup_files_checksum ||= dependencies_files_checksum(application.config._setup_dependencies)
        end
      end # GAPreSetupPatchDependencies
    end # Stage
  end # Build
end # Dapp
