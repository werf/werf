module SpecHelpers
  module Git
    extend ActiveSupport::Concern

    def repo_init!(git_dir: nil)
      FileUtils.mkdir_p git_dir unless git_dir.nil?
      git('init', git_dir: git_dir)
      git('commit --allow-empty -m init', git_dir: git_dir)
    end

    def repo_create_branch!(branch, git_dir: nil)
      git("branch #{branch}", git_dir: git_dir)
    end

    def repo_change_and_commit!(changefile='data.txt', changedata=random_string, branch: 'master', git_dir: nil)
      git("checkout #{branch}", git_dir: git_dir)
      changefile = File.join(git_dir, changefile)
      FileUtils.mkdir_p File.split(changefile)[0]
      File.write changefile, changedata
      repo_commit!(git_dir: git_dir)
    end

    def repo_commit!(git_dir: nil)
      git('add --all', git_dir: git_dir)
      unless git('diff --cached --quiet', returns: [0, 1], git_dir: git_dir).status.success?
        git('commit -m +', git_dir: git_dir)
      end
    end

    def repo_latest_commit(git_dir: nil, branch: 'master')
      git("rev-parse #{branch}", git_dir: git_dir).stdout.strip
    end

    def git(command, git_dir: nil, **kwargs)
      shellout "git #{git_dir.nil? ? '' : "-C #{git_dir}"} #{command}", **kwargs
    end

    included do
      before :all do
        shellout 'git config -l | grep "user.email" || git config --global user.email "dapp@flant.com"'
        shellout 'git config -l | grep "user.name" || git config --global user.name "Dapp Dapp"'
      end
    end
  end
end
