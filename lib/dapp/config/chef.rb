module Dapp
  module Config
    # Chef
    class Chef
      attr_reader :_modules
      attr_reader :_skip_modules
      attr_reader :_reset_modules

      def initialize
        @_modules = []
        @_skip_modules = []
        @_reset_modules = []
      end

      def module(*args)
        @_modules.concat(args)
      end

      def skip_module(*args)
        @_skip_modules.concat(args)
      end

      def reset_module(*args)
        @_reset_modules.concat(args)
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
