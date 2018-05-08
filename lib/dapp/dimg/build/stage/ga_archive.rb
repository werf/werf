module Dapp
  module Dimg
    module Build
      module Stage
        class GAArchive < GABase
          RESET_COMMIT_MESSAGES = ['[dapp reset]', '[reset dapp]'].freeze

          def initialize(dimg, next_stage)
            @prev_stage = BeforeInstallArtifact.new(dimg, self)
            super
          end

          def dependencies
            @dependencies ||= [dimg.git_artifacts.map(&:paramshash).join, reset_commits]
          end

          protected

          def reset_commits
            regex = Regexp.union(RESET_COMMIT_MESSAGES)
            dimg.git_artifacts.map { |git_artifact| git_artifact.repo.find_commit_id_by_message(regex) }.compact.sort.uniq
          end

          def apply_command_method
            :apply_archive_command
          end
        end # GAArchive
      end # Stage
    end # Build
  end # Dimg
end # Dapp
