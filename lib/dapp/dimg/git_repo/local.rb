module Dapp
  module Dimg
    module GitRepo
      class Local < Base
        attr_reader :path

        def initialize(manager, name, path)
          super(manager, name)
          self.path = path
        end

        def path=(path)
          @path ||= Pathname(Rugged::Repository.new(path).path)
        rescue Rugged::RepositoryError, Rugged::OSError => _e
          raise Error::Rugged, code: :local_git_repository_does_not_exist, data: { path: path }
        end

        def workdir_path
          Pathname(git.workdir)
        end

        def nested_git_directories_patches(paths: [], exclude_paths: [], **kwargs)
          patches(nil, nil, paths: paths, exclude_paths: exclude_paths, **kwargs).select do |patch|
            delta_new_file = patch.delta.new_file
            nested_git_repository_mode?(delta_new_file[:mode])
          end
        end

        # NOTICE: Параметры {from: nil, to: nil} можно указать только для Own repo.
        # NOTICE: Для Remote repo такой вызов не имеет смысла и это ошибка пользователя класса Remote.

        def submodules_params(commit, paths: [], exclude_paths: [])
          return super unless commit.nil?
          return []    unless File.file?((gitmodules_file_path = File.join(workdir_path, '.gitmodules')))

          submodules_params_base(File.read(gitmodules_file_path), paths: paths, exclude_paths: exclude_paths)
        end

        def ignore_patch?(patch, paths: [], exclude_paths: [])
          delta_new_file = patch.delta.new_file
          args = [delta_new_file[:path], paths: paths, exclude_paths: exclude_paths]
          if nested_git_repository_mode?(delta_new_file[:mode])
            !ignore_directory?(*args)
          else
            !ignore_path?(*args)
          end
        end

        def nested_git_repository_mode?(mode)
          mode == 0o040000
        end

        def diff(from, to, **kwargs)
          if from.nil? and to.nil?
            mid_commit = latest_commit
            diff_obj = super(nil, mid_commit, **kwargs)
            diff_obj.merge! git.lookup(mid_commit).diff_workdir(**kwargs)
            diff_obj
          elsif to.nil?
            git.lookup(from).diff_workdir(**kwargs)
          else
            super
          end
        end

        def latest_commit(_branch = nil)
          git.head.target_id
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError, TypeError => _e
          raise Error::Rugged, code: :commit_not_found_in_local_git_repository, data: { commit: commit }
        end
      end
    end
  end
end
