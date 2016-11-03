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
                          artifact_config = pass_to_default(ArtifactDimg.new("artifact-#{SecureRandom.hex(2)}", project: project))
                          artifact = Directive::Artifact.new(config: artifact_config)
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

      def check_dimg_directive_order(_directive)
      end

      def check_dimg_group_directive_order(_directive)
      end

      def pass_to_default(obj)
        super(obj).tap do |artifact_dimg|
          artifact_dimg.instance_variable_set(:@_artifact_dependencies, marshal_dup(_artifact_dependencies))
        end
      end
    end
  end
end
