module Dapp
  module Config
    # Directive
    module Directive
      # Artifact
      class Artifact < ArtifactBase
        attr_reader :_config

        def initialize(config:)
          @_config = config
          super()
        end

        def _export
          super do |export|
            export._before ||= @_before
            export._after ||= @_after
            export._config = _config
          end
        end

        # Export
        class Export < ArtifactBase::Export
          attr_accessor :_config
          attr_accessor :_before, :_after

          def before(stage)
            associate_validation!(:before, stage)
            @_before = stage
          end

          def after(stage)
            associate_validation!(:after, stage)
            @_after = stage
          end

          def not_associated?
            (_before || _after).nil?
          end

          protected

          def associate_validation!(type, stage)
            another = [:before, :after].find { |t| t != type }
            raise Error::Config, code: :stage_artifact_double_associate unless send("_#{another}").nil?
            raise Error::Config, code: :stage_artifact_not_supported_associated_stage unless [:install, :setup].include? stage
          end
        end
      end
    end
  end
end
