module Dapp
  module Config
    # Artifact
    class Artifact < Application
      attr_reader :_artifact_dependencies

      def initialize(parent)
        @_artifact_dependencies = []
        super
      end

      def artifact_depends_on(*args)
        @_artifact_dependencies.concat(args)
      end

      undef_method :app
    end
  end
end
