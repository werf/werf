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
            @dependencies ||= [dimg.git_artifacts.map(&:paramshash).join, reset_commits, dev_mode_dependencies]
          end

          def dev_mode_dependencies
            return unless dimg.dev_mode?
            dimg.git_artifacts.map(&:latest_commit)
          end

          protected

          def reset_commits
            regex = Regexp.union(RESET_COMMIT_MESSAGES)
            dimg.git_artifacts.map { |git_artifact| git_artifact.repo.find_commit_id_by_message(regex) }.sort.uniq.compact
          end

          def apply_command_method
            :apply_archive_command
          end
        end # GAArchive
      end # Stage
    end # Build
  end # Dimg
end # Dapp
