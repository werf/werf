module SpecHelpers
  module GitArtifact
    def git_artifact
      application.local_git_artifact_list.first
    end

    def stub_git_repo_own
      stub_instance(Dapp::GitRepo::Own) { |instance| instance.instance_variable_set(:@name, '') }
    end
  end
end
