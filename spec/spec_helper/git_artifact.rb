module SpecHelper
  module GitArtifact
    def git_artifact
      application.local_git_artifacts.first
    end

    def stub_git_repo_own
      stub_instance(Dapp::GitRepo::Own) do |instance|
        instance.instance_variable_set(:@name, '')
        allow(instance).to receive(:container_path) { '.git' }
      end
    end
  end
end
