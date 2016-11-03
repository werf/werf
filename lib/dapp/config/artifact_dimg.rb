module Dapp
  module Config
    class ArtifactDimg < Dimg
      attr_reader :_artifact_dependencies

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
