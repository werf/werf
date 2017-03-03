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

        def path
          @path ||= Rugged::Repository.discover(dimg.home_path.to_s).path
        rescue Rugged::RepositoryError => _e
          raise Error::Rugged, code: :local_git_repository_does_not_exist
        end

        def diff(from, to, **kwargs)
          if to.nil?
            git.lookup(from).diff_workdir
          else
            super
          end
        end

        def latest_commit(_branch = nil)
          git.head.target_id
        end

        def lookup_commit(commit)
          super
        rescue Rugged::OdbError => _e
          raise Error::Rugged, code: :commit_not_found_in_local_git_repository, data: { commit: commit }
        end
      end
    end
  end
end
