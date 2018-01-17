module Dapp
  module Dimg
    class Dimg
      module GitArtifact
        def git_artifacts
          [*local_git_artifacts, *remote_git_artifacts].compact
        end

        def local_git_artifacts
          @local_git_artifact_list ||= [].tap do |artifacts|
            break artifacts if (local_git_artifacts = Array(config._git_artifact._local)).empty?
            repo = GitRepo::Own.new(self)
            local_git_artifacts.map do |ga_config|
              artifacts.concat(generate_git_artifacts(repo, **ga_config._artifact_options))
            end
          end
        end

        def remote_git_artifacts
          @remote_git_artifact_list ||= [].tap do |artifacts|
            Array(config._git_artifact._remote).each do |ga_config|
              repo = GitRepo::Remote.new(self, ga_config._name, url: ga_config._url).tap { |r| r.fetch!(ga_config._branch) }
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
