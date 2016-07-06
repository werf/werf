module SpecHelpers
  module GitArtifact
    def git_artifact
      application.local_git_artifact_list.first
    end

    def stub_git_repo_own
      method_new = Dapp::GitRepo::Own.method(:new)

      git_repo = class_double(Dapp::GitRepo::Own).as_stubbed_const
      allow(git_repo).to receive(:new) do |*args, &block|
        method_new.call(*args, &block).tap { |instance| instance.instance_variable_set(:@name, '') }
      end
    end
  end
end
