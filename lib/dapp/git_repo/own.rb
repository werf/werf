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
      end

      def latest_commit(branch = nil)
        super(branch || 'HEAD')
      end
    end
  end
end
