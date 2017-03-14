module Dapp
  module Dimg
    module Config
      module Directive
        class GitArtifactLocal < ArtifactBase
          alias add export
          undef_method :export

          class Export < ArtifactBase::Export
            def stage_dependencies(&blk)
              @stage_dependencies ||= StageDependencies.new(&blk)
            end

            def _artifact_options
              super.merge(stages_dependencies: stage_dependencies.to_h)
            end

            class StageDependencies < Base
              STAGES = [:install, :setup, :before_setup, :build_artifact].freeze

              STAGES.each do |stage|
                define_method(stage) do |*glob|
                  if (globs = glob.map { |g| path_format(g) }).any? { |g| Pathname(g).absolute? }
                    raise Error::Config, code: :stages_dependencies_paths_relative_path_required, data: { stage: stage }
                  end
                  instance_variable_set(:"@#{stage}", globs)
                end

                define_method("_#{stage}") do
                  instance_variable_get(:"@#{stage}") || []
                end
              end

              def initialize(&blk)
                instance_eval(&blk) if block_given?
              end

              def to_h
                STAGES.map { |stage| [stage, send("_#{stage}")] }.to_h
              end
            end
          end
        end
      end
    end
  end
end
