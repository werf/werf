module SpecHelper
  module Git
    extend ActiveSupport::Concern

    def git_change_and_commit(changefile = 'data.txt', changedata = nil, git_dir: '.', msg: '+', binary: false)
      file_path = File.join(git_repo(git_dir: git_dir).workdir, changefile)
      unless (file_path_parts = File.split(file_path)).one?
        FileUtils.mkdir_p file_path_parts[0..-2].join('/')
      end

      changedata ||= binary ? random_binary_string : random_string
      File.open(file_path, 'w') { |f| f.write changedata }

      git_add_and_commit(changefile, git_dir: git_dir, msg: msg)
    end

    def git_add_and_commit(relative_path, git_dir: '.', msg: '+')
      git_add(relative_path, git_dir: git_dir)
      git_commit(git_dir: git_dir, msg: msg)
    end

    def git_rm_and_commit(relative_path, git_dir: '.', msg: '+')
      git_rm(relative_path, git_dir: git_dir)
      git_commit(git_dir: git_dir, msg: msg)
    end

    def git_add(relative_path, git_dir: '.')
      absolute_path = File.join(git_repo(git_dir: git_dir).workdir, relative_path)

      index = git_repo(git_dir: git_dir).index
      index.add path: relative_path,
                oid: (Rugged::Blob.from_workdir git_repo(git_dir: git_dir), relative_path),
                mode: File.symlink?(absolute_path) ? 40960 : File.stat(absolute_path).mode
    end

    def git_rm(relative_path, git_dir: '.')
      index = git_repo(git_dir: git_dir).index
      index.remove(relative_path)
    end

    def git_commit(git_dir: '.', msg: '+')
      index = git_repo(git_dir: git_dir).index

      commit_tree = index.write_tree git_repo(git_dir: git_dir)
      index.write

      commit_author = { email: 'dapp@test.com', name: 'Dapp', time: Time.now }
      Rugged::Commit.create git_repo(git_dir: git_dir),
                            author: commit_author,
                            committer: commit_author,
                            message: msg,
                            parents: git_repo(git_dir: git_dir).empty? ? [] : [ git_repo(git_dir: git_dir).head.target ],
                            tree: commit_tree,
                            update_ref: 'HEAD'
    end

    def git_checkout(branch, git_dir: '.')
      git_repo(git_dir: git_dir).checkout(branch)
    end

    def git_create_branch(branch, git_dir: '.')
      git_branches(git_dir: git_dir).create(branch, git_repo(git_dir: git_dir).head.target_id) unless git_branch_exist?(branch, git_dir: git_dir)
    end

    def git_branch_exist?(branch, git_dir: '.')
      git_branches(git_dir: git_dir).exist?(branch)
    end

    def git_branches(git_dir: '.')
      git_repo(git_dir: git_dir).branches
    end

    def git_latest_commit(git_dir: '.', branch: 'master')
      git_branches(git_dir: git_dir)[branch].target_id
    end

    def git_log(git_dir: '.')
      git_repo(git_dir: git_dir).head.log
    end

    def git_init(git_dir: '.')
      FileUtils.mkdir_p git_dir unless git_dir == '.'
      Rugged::Repository.init_at(git_dir)
      git_change_and_commit('README', git_dir: git_dir)
    end

    def git_repo(git_dir: '.')
      (@repo ||= {})[git_dir] ||= Rugged::Repository.new(git_dir)
    end
  end
end
