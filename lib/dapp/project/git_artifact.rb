module Dapp
  # Project
  class Project
    # GitArtifact
    module GitArtifact
      def local_git_artifact_exclude_paths(&blk)
        @local_git_artifact_exclude_paths ||= [].tap do |exclude_paths|
          yield exclude_paths if block_given?
        end
      end
    end # GitArtifact
  end # Project
end # Dapp
