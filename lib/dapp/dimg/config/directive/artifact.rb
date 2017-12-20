module Dapp
  module Dimg
    module Config
      module Directive
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

          class Export < ArtifactBase::Export
            attr_accessor :_config
            attr_accessor :_before, :_after

            def before(stage)
              sub_directive_eval do
                stage = stage.to_sym
                associate_validation!(:before, stage, @_before)
                @_before = stage
              end
            end

            def after(stage)
              sub_directive_eval do
                stage = stage.to_sym
                associate_validation!(:after, stage, @_after)
                @_after = stage
              end
            end

            def not_associated?
              (_before || _after).nil?
            end

            protected

            def associate_validation!(type, stage, _old_stage)
              conflict_type = [:before, :after].find { |t| t != type }
              conflict_stage = public_send("_#{conflict_type}")

              raise ::Dapp::Error::Config, code: :stage_artifact_not_supported_associated_stage,
                                           data: { stage: "#{type} #{stage.inspect}" } unless [:install, :setup].include? stage

              raise ::Dapp::Error::Config, code: :stage_artifact_double_associate,
                                           data: { stage: "#{type} #{stage.inspect}",
                                                   conflict_stage: "#{conflict_type} #{conflict_stage.inspect}" } if conflict_stage

              defined_stage = public_send("_#{type}")
              dapp.log_config_warning(
                desc: {
                  code: :stage_artifact_rewritten,
                  context: :warning,
                  data: { stage: "#{type} #{stage.inspect}",
                          conflict_stage: "#{type} #{defined_stage.inspect}" }
                }
              ) if defined_stage
            end
          end
        end
      end
    end
  end
end
