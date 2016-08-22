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
            super && !application.builder.install?
          end

          def dependencies
            prev_stage.prev_stage.dependencies # GAPreInstallPatchDependencies
          end

          def image
            super do |image|
              application.builder.install(image)
            end
          end
        end # Install
      end
    end # Stage
  end # Build
end # Dapp
