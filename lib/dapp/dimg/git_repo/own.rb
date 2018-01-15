module Dapp
  module Dimg
    module GitRepo
      class Own < Base
        def initialize(manager)
          super(manager, 'own')
        end

        def exclude_paths
          dapp.local_git_artifact_exclude_paths
        end

        def workdir_path
          Pathname(git.workdir)
        end

        def path
          @path ||= Pathname(Rugged::Repository.discover(dapp.path.to_s).path)
        rescue Rugged::RepositoryError => _e
          raise Error::Rugged, code: :local_git_repository_does_not_exist
        end

        # NOTICE: Параметры {from: nil, to: nil} можно указать только для Own repo.
        # NOTICE: Для Remote repo такой вызов не имеет смысла и это ошибка пользователя класса Remote.

        def submodules_params(commit, paths: [], exclude_paths: [])
          return super unless commit.nil?
          return []    unless File.file?((gitmodules_file_path = File.join(workdir_path, '.gitmodules')))

          submodules_params_base(File.read(gitmodules_file_path), paths: paths, exclude_paths: exclude_paths)
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

        def exist?
          super
        rescue Error::Rugged => e
          return false if e.net_status[:code] == :local_git_repository_does_not_exist
          raise
        end
      end
    end
  end
end
