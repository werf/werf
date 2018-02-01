module Dapp
  module Dimg
    module Config
      module Directive
        class ArtifactGroup < DimgGroup
          attr_reader :_name

          def initialize(name = nil, dapp:)
            super(dapp: dapp)
            @_name = name
          end

          def export(*args, &blk)
            _artifact_export(_artifact_config, *args, &blk)
          end

          def _artifact_config
            artifact_config_name = "artifact-#{[_name, SecureRandom.hex(2)].compact.join('-')}"
            pass_to(ArtifactDimg.new(artifact_config_name, dapp: dapp))
          end

          def _artifact_export(artifact_config, *args, &blk)
            artifact = Artifact.new(dapp: dapp, config: artifact_config)
            artifact.export(*args, &blk).tap do
              _export.concat artifact._export
            end
          end

          def _shell(&blk)
            @_shell ||= Shell::Artifact.new(dapp: dapp, &blk)
          end

          def _docker(&blk)
            @_docker ||= Docker::Artifact.new(dapp: dapp, &blk)
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
end
