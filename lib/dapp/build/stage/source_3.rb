module Dapp
  module Build
    module Stage
      class Source3 < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = InfraSetup.new(build, self)
          super
        end

        def name
          :source_3
        end

        def signature
          hashsum [prev_stage.signature,
                   app_setup_file,
                   *build.app_setup_commands,
                   *commit_list] # TODO chef
        end

        def git_artifact_signature
          hashsum [prev_stage.signature,
                   *build.app_setup_commands]
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
          build.build_path('.app_setup')
        end
      end # Source3
    end # Stage
  end # Build
end # Dapp
