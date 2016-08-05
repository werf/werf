module Dapp
  module Build
    module Stage
      # Source1
      class Source1 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = Source1Archive.new(application, self)
          super
        end

        def prev_source_stage
          prev_stage
        end

        protected

        def dependencies_checksum
          hashsum [super,
                   install_dependencies_files_checksum,
                   *application.builder.install_checksum]
        end

        private

        def install_dependencies_files_checksum
          @install_dependencies_files_checksum ||= dependencies_files_checksum(application.config._install_dependencies)
        end
      end # Source1
    end # Stage
  end # Build
end # Dapp
