module Dapp
  module Dimg
    module Config
      class ArtifactGroup < DimgGroup
        def _shell(&blk)
          @_shell ||= Directive::Shell::Artifact.new(dapp: dapp, &blk)
        end

        def _docker(&blk)
          @_docker ||= Directive::Docker::Artifact.new(dapp: dapp, &blk)
        end

        def _export
          @_export ||= []
        end

        undef :artifact
        undef :dimg
        undef :dimg_group

        protected

        def export(*args, &blk)
          _export.concat begin
            artifact_config = pass_to_default(
              ArtifactDimg.new(
                "artifact-#{SecureRandom.hex(2)}",
                dapp: dapp
              )
            )

            artifact = Directive::Artifact.new(dapp: dapp, config: artifact_config)
            artifact.send(:export, *args, &blk)

            artifact._export
          end
        end

        def check_dimg_directive_order(_directive)
        end

        def check_dimg_group_directive_order(_directive)
        end
      end
    end
  end
end
