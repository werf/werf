module Dapp
  module GitRepo
    # Own Git repo
    class Own < Base
      def initialize(application)
        super(application, 'own')
      end

      def dir_path
        @dir_path ||= Pathname(git("-C #{application.home_path} rev-parse --git-dir").stdout.strip).expand_path
      end

      def latest_commit(branch)
        super(branch || 'HEAD')
      end
    end
  end
end
