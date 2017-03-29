module Dapp
  module Dimg
    module Config
      module Directive
        class ArtifactDimg < Dimg
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
end
