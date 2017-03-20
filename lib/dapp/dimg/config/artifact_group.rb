module Dapp
  module Dimg
    module Config
      class ArtifactGroup < DimgGroup
        def export(*args, &blk)
          artifact_config = pass_to_default(ArtifactDimg.new("artifact-#{SecureRandom.hex(2)}", dapp: dapp))
          artifact = Directive::Artifact.new(dapp: dapp, config: artifact_config)
          artifact.export(*args, &blk).tap do
            _export.concat artifact._export
          end
        end

        def _shell(&blk)
          @_shell ||= Directive::Shell::Artifact.new(dapp: dapp, &blk)
        end

        def _docker(&blk)
          @_docker ||= Directive::Docker::Artifact.new(dapp: dapp, &blk)
        end

        def _export
          @_export ||= []
        end

        def validate!
          _export.each(&:validate!)
        end

        undef :artifact
        undef :dimg
        undef :dimg_group

        protected

        def check_dimg_directive_order(_directive)
        end

        def check_dimg_group_directive_order(_directive)
        end
      end
    end
  end
end
