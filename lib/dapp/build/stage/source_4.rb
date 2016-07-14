module Dapp
  module Build
    module Stage
      # Source4
      class Source4 < SourceBase
        MAX_PATCH_SIZE = 1024 * 1024

        def initialize(application, next_stage)
          @prev_stage = AppSetup.new(application, self)
          super
        end

        def next_source_stage
          next_stage
        end

        def dependencies_checksum
          hashsum [super, (changes_size_since_source3 / MAX_PATCH_SIZE).to_i]
        end

        private

        def changes_size_since_source3
          application.git_artifacts.map do |git_artifact|
            git_artifact.patch_size(prev_source_stage.layer_commit(git_artifact), git_artifact.latest_commit)
          end.reduce(0, :+)
        end
      end # Source4
    end # Stage
  end # Build
end # Dapp
