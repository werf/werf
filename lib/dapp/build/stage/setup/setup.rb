module Dapp
  module Build
    module Stage
      module SetupGroup
        # Setup
        class Setup < Base
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPreSetupPatch.new(application, self)
            super
          end

          def empty?
            super && !application.builder.setup?
          end

          def dependencies
            [setup_dependencies_files_checksum, application.builder.setup_checksum]
          end

          def prepare_image
            super
            application.builder.setup(image)
          end

          private

          def setup_dependencies_files_checksum
            @setup_files_checksum ||= dependencies_files_checksum(application.config._setup_dependencies)
          end
        end # Setup
      end
    end # Stage
  end # Build
end # Dapp
