module Dapp
  module Config
    module Directive
      class Artifact < Directive::GitArtifactLocal
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

        protected

        class Export < Directive::GitArtifactLocal::Export
          attr_accessor :_config
          attr_accessor :_before, :_after

          def before(stage)
            validation(:before, stage)
            @_before = stage
          end

          def after(stage)
            validation(:after, stage)
            @_after = stage
          end

          protected

          def validation(type, stage)
            another = [:before, :after].find { |t| t != type }
            raise Error::Config, code: :stage_artifact_double_associate unless send("_#{another}").nil?
            raise Error::Config, code: :stage_artifact_not_supported_associated_stage unless [:install, :setup].include? stage
          end
        end
      end
    end
  end
end
