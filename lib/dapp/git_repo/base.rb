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

      def diff(from = nil, to, **kwargs)
        if from.nil?
          Rugged::Tree.diff(git_bare, nil, to, **kwargs)
        else
          lookup_commit(from).diff(lookup_commit(to), **kwargs)
        end
      end

      def patches(from = nil, to, exclude_paths: [], **kwargs)
        diff(from, to, **kwargs).patches.select do |patch|
          !exclude_paths.any? { |p| check_path?(patch.delta.new_file[:path], p) }
        end
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

      def find_commit_id_by_message(regex)
        walker.each do |commit|
          next unless commit.message =~ regex
          return commit.oid
        end
      end

      def walker
        walker = Rugged::Walker.new(git_bare)
        walker.push(git_bare.head.target_id)
        walker
      end

      def lookup_object(oid)
        git_bare.lookup(oid)
      end

      def lookup_commit(commit)
        git_bare.lookup(commit)
      end

      protected

      def git
        @git ||= Rugged::Repository.new(path)
      end

      private

      def check_path?(path, format)
        path_parts = path.split('/')
        checking_path = nil

        until path_parts.empty?
          checking_path = [checking_path, path_parts.shift].compact.join('/')
          return true if File.fnmatch(format, checking_path)
        end
        false
      end
    end
  end
end
