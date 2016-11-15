module Dapp
  module Build
    module Stage
      # BuildArtifact
      class BuildArtifact < Base
        def initialize(dimg)
          @prev_stage = GAArtifactPatch.new(dimg, self)
          @dimg = dimg
        end

        def empty?
          !dimg.builder.build_artifact?
        end

        def context
          [artifact_dependencies_files_checksum, builder_checksum]
        end

        def builder_checksum
          dimg.builder.build_artifact_checksum
        end

        def prepare_image
          super
          dimg.builder.build_artifact(image)
        end

        private

        def artifact_dependencies_files_checksum
          @artifact_files_checksum ||= dependencies_files_checksum(dimg.config._artifact_dependencies)
        end
      end # BuildArtifact
    end # Stage
  end # Build
end # Dapp
