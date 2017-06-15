module Dapp
  module Dimg
    class Dimg
      module GitArtifact
        def git_artifacts
          [*local_git_artifacts, *remote_git_artifacts].compact
        end

        def local_git_artifacts
          @local_git_artifact_list ||= begin
            repo = GitRepo::Own.new(self)
            Array(config._git_artifact._local).map do |ga_config|
              ::Dapp::Dimg::GitArtifact.new(repo, **ga_config._artifact_options)
            end
          end
        end

        def remote_git_artifacts
          @remote_git_artifact_list ||= begin
            repos = {}
            Array(config._git_artifact._remote).map do |ga_config|
              repo_key = [ga_config._url, ga_config._branch]
              repos[repo_key] ||= GitRepo::Remote.new(self, ga_config._name, url: ga_config._url).tap { |repo| repo.fetch!(ga_config._branch) }
              ::Dapp::Dimg::GitArtifact.new(repos[repo_key], **ga_config._artifact_options)
            end
          end
        end
      end # GitArtifact
    end # Mod
  end # Dimg
end # Dapp
