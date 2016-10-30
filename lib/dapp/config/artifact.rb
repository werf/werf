module Dapp
  module Config
    class Artifact < Dimg
      attr_reader :_artifact_dependencies

      def initialize(project)
        @_artifact_dependencies = []
        super(project: project)
      end

      def artifact_depends_on(*args)
        @_artifact_dependencies.concat(args)
      end
    end
  end
end
