module Dapp
  module Config
    module Directive
      module Shell
        class Artifact < Base
          attr_reader :_build_artifact
          attr_reader :_build_artifact_cache_version

          def initialize
            super
            @_build_artifact = []
          end

          def build_artifact(*args, cache_version: nil)
            @_build_artifact.concat(args)
            @_build_artifact_cache_version = cache_version
          end

          def reset_build_artifact
            @_build_artifact = []
          end

          def reset_all
            super
            reset_build_artifact
          end

          protected

          def empty?
            super && @_build_artifact.empty?
          end
        end
      end
    end
  end
end
