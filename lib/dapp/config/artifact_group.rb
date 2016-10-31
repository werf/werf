module Dapp
  module Config
    class ArtifactGroup < DimgGroup
      attr_reader :_artifact_dependencies, :_export

      def initialize(project:)
        @_artifact_dependencies = []
        @_export = []

        super(project: project)
      end

      def artifact_depends_on(*args)
        @_artifact_dependencies.concat(args)
      end

      def export(*args, &blk)
        @_export.concat begin
          artifact = Directive::Artifact.new(config: clone)
          artifact.export(*args, &blk)
          artifact._export
        end
      end

      def _shell(&blk)
        @_shell ||= Directive::Shell::Artifact.new(&blk)
      end

      def _docker(&blk)
        @_docker ||= Directive::Docker::Artifact.new(&blk)
      end

      undef :artifact
      undef :dimg
      undef :dimg_group

      protected

      def pass_to_default(obj)
        super(obj).tap do |artifact_group|
          artifact_group.instance_variable_set(:@_artifact_dependencies, marshal_dup(_artifact_dependencies))
        end
      end

      def clone
        pass_to_default(self.class.new(project: project))
      end
    end
  end
end
