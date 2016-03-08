module Dapper
  module GitRepo
    # Base class for any Git repo (remote, gitkeeper, etc)
    class Base
      attr_reader :builder
      attr_reader :name
      attr_reader :su

      def initialize(builder, name, build_path: nil)
        @builder = builder
        @name = name
        @build_path = build_path || []
      end

      def build_path(*paths)
        builder.build_path(*@build_path, *paths)
      end

      def dir_path
        build_path "#{name}.git"
      end

      def git(command, **kwargs)
        builder.shellout "git #{command}", **kwargs
      end

      def git_bare(command, **kwargs)
        git "--git-dir=#{dir_path} #{command}", **kwargs
      end

      def commit_at(commit)
        Time.at Integer git_bare("show -s --format=%ct #{commit}").stdout.strip
      end

      def latest_commit(branch = 'master')
        git_bare("rev-parse #{branch}").stdout.strip
      end

      def cleanup!
        lock do
          FileUtils.rm_rf dir_path
        end
      end

      def lock(**kwargs, &block)
        builder.filelock(build_path("#{name}.lock"), error_message: "Repository #{name} in use! Try again later.", **kwargs, &block)
      end
    end
  end
end
