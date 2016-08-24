module Dapp
  module Build
    module Stage
      # GAArchive
      class GAArchive < GABase
        def initialize(application, next_stage)
          @prev_stage = GAArchiveDependencies.new(application, self)
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
          :archive_apply_command
        end
      end # GAArchive
    end # Stage
  end # Build
end # Dapp
