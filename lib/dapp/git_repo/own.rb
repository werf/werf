module Dapp
  module GitRepo
    # Own Git repo
    class Own < Base
      def initialize(dimg)
        super(dimg, 'own')
      end

      def container_path
        dimg.container_dapp_path('own', "#{name}.git")
      end

      def path
        @path ||= Rugged::Repository.discover(dimg.home_path.to_s).path
      rescue Rugged::RepositoryError => _e
        raise Error::Rugged, code: :local_git_repository_does_not_exist
      end

      def latest_commit(branch = nil)
        super(branch || 'HEAD')
      end

      def lookup_commit(commit)
        super
      rescue Rugged::OdbError => _e
        raise Error::Rugged, code: :commit_not_found_in_local_git_repository, data: { commit: commit }
      end
    end
  end
end
