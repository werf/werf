module SpecHelpers
  module Git
    extend ActiveSupport::Concern

    def repo
      @repo ||= Dapp::GitRepo::Chronicler.new(application, repo_name)
    end

    def repo_name
      raise
    end

    def repo_create_branch(branch)
      shellout "git branch #{branch}", cwd: repo.name
    end

    def repo_change_and_commit(changefile: 'data.txt', changedata: random_string, branch: 'master')
      shellout "git checkout #{branch}", cwd: repo.name
      changefile = File.join(repo.name, changefile)
      FileUtils.mkdir_p File.split(changefile)[0]
      File.write changefile, changedata
      repo.commit!
    end

    def repo_path
      Pathname('/tmp/dapp/hello')
    end

    def commit!
      git 'add --all'
      unless git('diff --cached --quiet', returns: [0, 1]).status.success?
        git 'commit -m +'
      end
    end

    def git(command, **kwargs)
      shellout "git -C #{repo_path} #{command}", **kwargs
    end

    included do
      before :all do
        shellout 'git config -l | grep "user.email" || git config --global user.email "dapp@flant.com"'
        shellout 'git config -l | grep "user.name" || git config --global user.name "Dapp Dapp"'
      end
    end
  end
end
