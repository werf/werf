module SpecHelpers
  module Git
    extend ActiveSupport::Concern

    def repo_init
      git 'init'
      git 'commit --allow-empty -m init'
      repo_change_and_commit('README.md', 'Hello')
    end

    def repo_create_branch(branch)
      shellout "git branch #{branch}"
    end

    def repo_change_and_commit(changefile='data.txt', changedata=random_string, branch: 'master')
      shellout "git checkout #{branch}"
      changefile = File.join(changefile)
      FileUtils.mkdir_p File.split(changefile)[0]
      File.write changefile, changedata
      repo_commit!
    end

    def repo_commit!
      git 'add --all'
      unless git('diff --cached --quiet', returns: [0, 1]).status.success?
        git 'commit -m +'
      end
    end

    def git(command, **kwargs)
      shellout "git #{command}", **kwargs
    end

    included do
      before :all do
        shellout 'git config -l | grep "user.email" || git config --global user.email "dapp@flant.com"'
        shellout 'git config -l | grep "user.name" || git config --global user.name "Dapp Dapp"'
      end
    end
  end
end
