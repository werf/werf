module Dapp
  module Build
    module Stage
      # Source1Archive
      class Source1Archive < SourceBase
        def initialize(application, next_stage)
          @prev_stage = Source1ArchiveDependencies.new(application, self)
          super
        end

        def prev_source_stage
          nil
        end

        def next_source_stage
          next_stage.next_stage
        end

        protected

        def apply_command_method
          :archive_apply_command
        end
      end # Source1Archive
    end # Stage
  end # Build
end # Dapp
