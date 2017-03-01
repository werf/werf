module Dapp
  # Dimg
  class Dimg
    # GitArtifact
    module GitArtifact
      def git_artifacts
        [*local_git_artifacts, *remote_git_artifacts].compact
      end

      def local_git_artifacts
        @local_git_artifact_list ||= Array(config._git_artifact._local).map do |ga_config|
          repo = GitRepo::Own.new(self)
          ::Dapp::GitArtifact.new(repo, **ga_config._artifact_options)
        end
      end

      def remote_git_artifacts
        @remote_git_artifact_list ||= Array(config._git_artifact._remote).map do |ga_config|
          repo = GitRepo::Remote.new(self, ga_config._name, url: ga_config._url)
          repo.fetch!(ga_config._branch)
          ::Dapp::GitArtifact.new(repo, **ga_config._artifact_options)
        end
      end
    end # GitArtifact
  end # Dimg
end # Dapp
