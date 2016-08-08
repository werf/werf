module Dapp
  module Build
    module Stage
      # Source3Dependencies
      class Source3Dependencies < SourceDependenciesBase
        def initialize(application, next_stage)
          @prev_stage = InfraSetup.new(application, self)
          super
        end

        def dependencies
          [setup_dependencies_files_checksum, application.builder.setup_checksum]
        end

        private

        def setup_dependencies_files_checksum
          @setup_files_checksum ||= dependencies_files_checksum(application.config._setup_dependencies)
        end
      end # Source3Dependencies
    end # Stage
  end # Build
end # Dapp
