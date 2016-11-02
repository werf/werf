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
        @path ||= Pathname(git("-C #{dimg.home_path} rev-parse --git-dir").stdout.strip).expand_path
      end

      def latest_commit(branch = nil)
        super(branch || 'HEAD')
      end
    end
  end
end
