module Dapp
  module GitRepo
    # Own Git repo
    class Own < Base
      def initialize(application)
        super(application, 'own')
      end

      def container_path
        application.container_dapp_path('own', "#{name}.git")
      end

      def path
        @path ||= Pathname(git("-C #{application.home_path} rev-parse --git-dir").stdout.strip).expand_path
      end

      def latest_commit(branch = nil)
        super(branch || 'HEAD')
      end
    end
  end
end
