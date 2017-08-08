module Dapp
  module Dimg
    module GitRepo
      class Own < Base
        def initialize(dimg)
          super(dimg, 'own')
        end

        def exclude_paths
          dimg.dapp.local_git_artifact_exclude_paths
        end

        def workdir_path
          Pathname(git.workdir)
        end

        def path
          @path ||= Pathname(Rugged::Repository.discover(dimg.home_path.to_s).path)
        rescue Rugged::RepositoryError => _e
          raise Error::Rugged, code: :local_git_repository_does_not_exist
        end

        # NOTICE: Параметры {from: nil, to: nil} можно указать только для Own repo.
        # NOTICE: Для Remote repo такой вызов не имеет смысла и это ошибка пользователя класса Remote.

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
