module Dapp
  module GitRepo
    # Own Git repo
    class Own < Base
      def initialize(builder, **kwargs)
        super(builder, 'own', **kwargs)
      end

      def dir_path
        @dir_path ||= git("-C #{builder.home_path} rev-parse --git-dir").stdout.strip
      end
    end
  end
end
