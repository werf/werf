module Dapp
  module Dimg
    module Build
      module Stage
        class GAArchive < GABase
          def initialize(dimg, next_stage)
            @prev_stage = GAArchiveDependencies.new(dimg, self)
            super
          end

          def prev_g_a_stage
            nil
          end

          def next_g_a_stage
            next_stage.next_stage # GAPreInstallPatch
          end

          protected

          def apply_command_method
            :apply_archive_command
          end
        end # GAArchive
      end # Stage
    end # Build
  end # Dimg
end # Dapp
