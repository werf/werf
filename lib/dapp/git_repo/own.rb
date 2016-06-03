module Dapp
  module GitRepo
    # Own Git repo
    class Own < Base
      def initialize(builder, **kwargs)
        super(builder, 'own', **kwargs)
      end

      def dir_path
        @dir_path ||= Pathname(git("-C #{builder.home_path} rev-parse --git-dir").stdout.strip).expand_path
      end
    end
  end
end
