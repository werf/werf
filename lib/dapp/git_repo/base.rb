module Dapp
  module GitRepo
    # Base class for any Git repo (remote, gitkeeper, etc)
    class Base
      attr_reader :application
      attr_reader :name

      def initialize(application, name)
        @application = application
        @name = name
      end

      def container_path
        application.container_tmp_path "#{name}.git"
      end

      def path
        application.tmp_path "#{name}.git"
      end

      def git_bare(command, **kwargs)
        git "--git-dir=#{path} #{command}", **kwargs
      end

      def commit_at(commit)
        Time.at Integer git_bare("show -s --format=%ct #{commit}").stdout.strip
      end

      def latest_commit(branch)
        git_bare("rev-parse #{branch}").stdout.strip
      end

      def exist_in_commit?(path, commit)
        git_bare("cat-file -e #{commit}:#{path}", returns: [0, 128]).status.success?
      end

      def cleanup!
      end

      def branch
        git_bare('rev-parse --abbrev-ref HEAD').stdout.strip
      end

      protected

      def git(command, **kwargs)
        application.project.shellout!("git #{command}", **kwargs)
      end
    end
  end
end
