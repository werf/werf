module Dapp
  module Build
    module Stage
      # Source3
      class Source3 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = InfraSetup.new(application, self)
          super
        end

        protected

        def dependencies_checksum
          hashsum [prev_stage.signature,
                   app_setup_files_checksum,
                   *application.builder.app_setup_checksum]
        end

        private

        def app_setup_files_checksum
          @app_setup_files_checksum ||= dependency_files_checksum(application.config._app_setup_dependencies)
        end
      end # Source3
    end # Stage
  end # Build
end # Dapp
