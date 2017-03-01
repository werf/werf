module Dapp
  module Dimg
    module Config
      # ArtifactDimg
      class ArtifactDimg < Dimg
        def _artifact_dependencies
          @_artifact_dependencies ||= []
        end

        def validate_scratch!
        end

        def validate_artifacts_artifacts!
        end

        def validated_artifacts
          _git_artifact._local + _git_artifact._remote
        end
      end
    end
  end
end
