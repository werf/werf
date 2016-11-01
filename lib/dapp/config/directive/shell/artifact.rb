module Dapp
  module Config
    module Directive
      module Shell
        class Artifact < Dimg
          attr_reader :_build_artifact
          stage_command_generator(:build_artifact)

          def empty?
            super && _build_artifact_command.empty?
          end
        end
      end
    end
  end
end
