module Dapp
  module Config
    # ArtifactGroup
    class ArtifactGroup < DimgGroup
      attr_reader :_artifact_dependencies, :_export

      def initialize(project:)
        @_artifact_dependencies = []
        @_export = []

        super(project: project)
      end

      def _shell(&blk)
        @_shell ||= Directive::Shell::Artifact.new(project: project, &blk)
      end

      def _docker(&blk)
        @_docker ||= Directive::Docker::Artifact.new(project: project, &blk)
      end

      undef :artifact
      undef :dimg
      undef :dimg_group

      protected

      def artifact_depends_on(*args)
        @_artifact_dependencies.concat(args)
      end

      def export(*args, &blk)
        @_export.concat begin
                          artifact_config = pass_to_default(ArtifactDimg.new("artifact-#{SecureRandom.hex(2)}", project: project))
                          artifact = Directive::Artifact.new(project: project, config: artifact_config)
                          artifact.send(:export, *args, &blk)
                          artifact._export
                        end
      end

      def check_dimg_directive_order(_directive)
      end

      def check_dimg_group_directive_order(_directive)
      end

      def pass_to_default(obj)
        super(obj).tap do |artifact_dimg|
          artifact_dimg.instance_variable_set(:@_artifact_dependencies, _artifact_dependencies.dup)
        end
      end
    end
  end
end
