module Dapp
  module Build
    module Stage
      class Source3 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = InfraSetup.new(application, self)
          super
        end

        protected

        def dependencies_checksum
          hashsum [prev_stage.signature,
                   app_setup_file,
                   *application.builder.app_setup_checksum]
        end

        private

        def app_setup_file
          @app_setup_file ||= begin
            File.read(app_setup_file_path) if app_setup_file?
          end
        end

        def app_setup_file?
          File.exist?(app_setup_file_path)
        end

        def app_setup_file_path
          application.build_path('.app_setup')
        end
      end # Source3
    end # Stage
  end # Build
end # Dapp
