module Dapp
  module Build
    module Stage
      class Source4 < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = AppSetup.new(build, self)
          super
        end

        def next_source_stage
          next_stage
        end

        def signature
          if patches_size_valid?
            hashsum prev_stage.signature
          else
            hashsum [dependencies_checksum, *commit_list]
          end
        end

        private

        def patches_size_valid?
          build.git_artifact_list.all? { |git_artifact| patch_size_valid?(git_artifact) }
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
