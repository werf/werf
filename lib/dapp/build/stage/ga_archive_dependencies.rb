module Dapp
  module Build
    module Stage
      # GAArchiveDependencies
      class GAArchiveDependencies < GADependenciesBase
        RESET_COMMIT_MESSAGES = ['[dapp reset]', '[reset dapp]']

        def initialize(dimg, next_stage)
          @prev_stage = BeforeInstallArtifact.new(dimg, self)
          super
        end

        def dependencies
          [dimg.git_artifacts.map(&:paramshash).join, reset_commits.sort.uniq.compact]
        end

        protected

        def reset_commits
          regex = Regexp.union(RESET_COMMIT_MESSAGES)
          dimg.git_artifacts.map { |git_artifact| git_artifact.repo.find_commit_id_by_message(regex) }
        end
      end # GAArchiveDependencies
    end # Stage
  end # Build
end # Dapp
