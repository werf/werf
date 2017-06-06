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
          @git ||= Rugged::Repository.new(path.to_s, **kwargs)
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
end
