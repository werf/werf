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
          hashsum [prev_stage.signature,
                   app_install_files_checksum,
                   *application.builder.app_install_checksum]
        end

        private

        def app_install_files_checksum
          @app_install_files_checksum ||= dependency_files_checksum(application.config._app_install_dependencies)
        end
      end # Source1
    end # Stage
  end # Build
end # Dapp
