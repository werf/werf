module Dapp
  module Config
    # Directive
    module Directive
      # Artifact
      class Artifact < ArtifactBase
        attr_reader :_config

        def initialize(config:, **kwargs, &blk)
          @_config = config

          super(**kwargs, &blk)
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

          def not_associated?
            (_before || _after).nil?
          end

          protected

          def before(stage)
            stage = stage.to_sym
            associate_validation!(:before, stage, @_before)
            @_before = stage
          end

          def after(stage)
            stage = stage.to_sym
            associate_validation!(:after, stage, @_after)
            @_after = stage
          end

          def associate_validation!(type, stage, old_stage)
            conflict_type = [:before, :after].find { |t| t != type }
            conflict_stage = send("_#{conflict_type}")

            raise Error::Config, code: :stage_artifact_not_supported_associated_stage,
                                 data: { stage: "#{type} #{stage.inspect}" }  unless [:install, :setup].include? stage

            raise Error::Config, code: :stage_artifact_double_associate,
                                 data: { stage: "#{type} #{stage.inspect}",
                                         conflict_stage: "#{conflict_type} #{conflict_stage.inspect}" } if conflict_stage

            defined_stage = send("_#{type}")
            dapp.log_config_warning(desc: {
              code: :stage_artifact_rewritten,
              context: :warning,
              data: { stage: "#{type} #{stage.inspect}",
                      conflict_stage: "#{type} #{defined_stage.inspect}" }
            }) if defined_stage
          end
        end
      end
    end
  end
end
