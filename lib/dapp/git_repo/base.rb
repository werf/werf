module Dapp
  module GitRepo
    # Base class for any Git repo (remote, gitkeeper, etc)
    class Base
      attr_reader :dimg
      attr_reader :name

      def initialize(dimg, name)
        @dimg = dimg
        @name = name
      end

      def container_path
        dimg.container_tmp_path "#{name}.git"
      end

      def path
        dimg.tmp_path("#{name}.git").to_s
      end

      def git_bare
        @git_bare ||= Rugged::Repository.new(path, bare: true)
      end

      def old_git_bare(command, **kwargs)
        old_git "--git-dir=#{path} #{command}", **kwargs
      end

      def commit_at(commit)
        git_bare.lookup(commit).time.to_i
      end

      def latest_commit(branch)
        return git_bare.head.target_id if branch == 'HEAD'
        git_bare.branches[branch].target_id
      end

      def cleanup!
      end

      def branch
        git_bare.head.name.sub(/^refs\/heads\//, '')
      end

      protected

      def git
        @git ||= Rugged::Repository.new(path)
      end

      def old_git(command, **kwargs)
        dimg.system_shellout! "#{dimg.project.git_bin} #{command}", **kwargs
      end
    end
  end
end
