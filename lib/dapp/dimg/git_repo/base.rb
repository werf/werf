module Dapp
  module Dimg
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

        # FIXME: Убрать логику исключения путей exclude_paths из данного класса,
        # FIXME: т.к. большинство методов не поддерживают инвариант
        # FIXME "всегда выдавать данные с исключенными путями".
        # FIXME: Например, метод diff выдает данные без учета exclude_paths.
        # FIXME: Лучше перенести фильтрацию в GitArtifact::diff_patches.
        # FIXME: ИЛИ обеспечить этот инвариант, но это ограничит в возможностях
        # FIXME: использование Rugged извне этого класса и это более сложный путь.
        # FIXME: Лучше сейчас убрать фильтрацию, а добавить ее когда наберется достаточно
        # FIXME: примеров использования.

        def patches(from, to, exclude_paths: [], **kwargs)
          diff(from, to, **kwargs).patches.select do |patch|
            !exclude_paths.any? { |p| check_path?(patch.delta.new_file[:path], p) }
          end
        end

        def entries(commit, paths: [], exclude_paths: [])
          [].tap do |entries|
            lookup_commit(commit).tree.walk(:preorder) do |root, entry|
              fullpath = File.join(root, entry[:name]).reverse.chomp('/').reverse

              is_exclude_path = exclude_paths.any? { |p| check_path?(fullpath, p) }
              is_include_path = begin
                paths.empty? ||
                  paths.any? { |p| check_path?(fullpath, p) || check_subpath?(fullpath, p) }
              end

              next false if is_exclude_path || !is_include_path

              entries << [root, entry]
            end
          end
        end

        def diff(from, to, **kwargs)
          if to.nil?
            raise "Workdir diff not supported for #{self.class}"
          elsif from.nil?
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
        rescue Rugged::ReferenceError => e
          raise Error::Rugged, code: :git_repository_reference_error, data: { name: name, message: e.message.downcase }
        end

        def commit_at(commit)
          lookup_commit(commit).time.to_i
        end

        def find_commit_id_by_message(regex)
          walker.each do |commit|
            msg = commit.message.encode('UTF-8', invalid: :replace, undef: :replace)
            return commit.oid if msg =~ regex
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
          @git ||= Rugged::Repository.new(path.to_s, **kwargs)
        end

        private

        def check_path?(path, format)
          path_checker(path) do |checking_path|
            File.fnmatch(format, checking_path)
          end
        end

        def check_subpath?(path, format)
          path_checker(format) do |checking_path|
            File.fnmatch(checking_path, path)
          end
        end

        def path_checker(path)
          path_parts = path.split('/')
          checking_path = nil

          until path_parts.empty?
            checking_path = [checking_path, path_parts.shift].compact.join('/')
            return true if yield checking_path
          end
          false
        end
      end
    end
  end
end
