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

      def exclude_paths
        []
      end

      def patches(from, to, exclude_paths: [], **kwargs)
        diff(from, to, **kwargs).patches.select do |patch|
          !exclude_paths.any? { |p| check_path?(patch.delta.new_file[:path], p) }
        end
      end

      def diff(from, to, **kwargs)
        if from.nil?
          Rugged::Tree.diff(git, nil, to, **kwargs)
        else
          lookup_commit(from).diff(lookup_commit(to), **kwargs)
        end
      end

      def commit_exists?(commit)
        git.exists?(commit)
      end

      def latest_commit(_branch)
        raise
      end

      def branch
        git.head.name.sub(/^refs\/heads\//, '')
      end

      def commit_at(commit)
        lookup_commit(commit).time.to_i
      end

      def find_commit_id_by_message(regex)
        walker.each do |commit|
          next unless commit.message =~ regex
          return commit.oid
        end
      end

      def walker
        walker = Rugged::Walker.new(git)
        walker.push(git.head.target_id)
        walker
      end

      def lookup_object(oid)
        git.lookup(oid)
      end

      def lookup_commit(commit)
        git.lookup(commit)
      end

      protected

      def git(**kwargs)
        @git ||= Rugged::Repository.new(path, **kwargs)
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
