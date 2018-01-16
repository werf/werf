module Dapp
  module Dimg
    class Dimg
      module GitArtifact
        def git_artifacts
          [*local_git_artifacts, *remote_git_artifacts].compact
        end

        def local_git_artifacts
          @local_git_artifact_list ||= [].tap do |artifacts|
            repo = GitRepo::Own.new(self)
            Array(config._git_artifact._local).map do |ga_config|
              artifacts.concat(generate_git_artifacts(repo, **ga_config._artifact_options))
            end
          end
        end

        def remote_git_artifacts
          @remote_git_artifact_list ||= [].tap do |artifacts|
            Array(config._git_artifact._remote).each do |ga_config|
              repo = GitRepo::Remote.get_or_init(self, ga_config._name, url: ga_config._url, branch: ga_config._branch)
              artifacts.concat(generate_git_artifacts(repo, **ga_config._artifact_options))
            end
          end
        end

        def generate_git_artifacts(repo, **git_artifact_options)
          [].tap do |artifacts|
            artifacts << (artifact = ::Dapp::Dimg::GitArtifact.new(repo, **git_artifact_options))
            artifacts.concat(generate_git_submodules_artifacts(artifact))
          end
        end

        def generate_git_submodules_artifacts(artifact)
          [].tap do |artifacts|
            artifacts.concat(submodules_artifacts = artifact.submodules_artifacts)
            artifacts.concat(submodules_artifacts.map(&method(:generate_git_submodules_artifacts)).flatten)
          end
        end
      end # GitArtifact
    end # Mod
  end # Dimg
end # Dapp
