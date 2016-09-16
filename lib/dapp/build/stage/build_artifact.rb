module Dapp
  module Build
    module Stage
      # BuildArtifact
      class BuildArtifact < Base
        def initialize(application)
          @prev_stage = GAArtifactPatch.new(application, self)
          @application = application
        end

        def empty?
          !application.builder.build_artifact?
        end

        def context
          [artifact_dependencies_files_checksum, application.builder.build_artifact_checksum]
        end

        def prepare_image
          super
          application.builder.build_artifact(image)
        end

        private

        def artifact_dependencies_files_checksum
          @artifact_files_checksum ||= dependencies_files_checksum(application.config._artifact_dependencies)
        end
      end # GAArchive
    end # Stage
  end # Build
end # Dapp
