module Dapp
  module Config
    module Directive
      module Shell
        class Artifact < Dimg
          attr_reader :_build_artifact
          stage_command_generator(:build_artifact)
        end
      end
    end
  end
end
