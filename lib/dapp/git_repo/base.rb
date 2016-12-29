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

      def diff(from, to, **kwargs)
        lookup_commit(from).diff(lookup_commit(to), **kwargs)
      end

      def commit_at(commit)
        lookup_commit(commit).time.to_i
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

      def file_exist_in_tree?(tree, paths)
        path = paths.shift
        paths.empty? ?
          tree.each { |obj| return true if obj[:name] == path } :
          tree.each_tree { |tree| return file_exist_in_tree?(tree, paths) if tree[:name] == path }
        false
      end

      def lookup_commit(commit)
        git_bare.lookup(commit)
      end

      protected

      def git
        @git ||= Rugged::Repository.new(path)
      end
    end
  end
end
