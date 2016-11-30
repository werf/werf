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
        @path ||= begin
          git_repo_path = Pathname(git("-C #{dimg.home_path} rev-parse --git-dir").stdout.strip)
          if git_repo_path.relative?
            File.join(dimg.home_path, git_repo_path)
          else
            git_repo_path
          end
        end
      end

      def latest_commit(branch = nil)
        super(branch || 'HEAD')
      end
    end
  end
end
