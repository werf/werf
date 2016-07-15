module Dapp
  module Config
    # Chef
    class Chef
      attr_reader :_modules

      def initialize
        @_modules = []
      end

      def module(*args)
        @_modules.concat(args)
      end

      def to_h
        {
          modules: _modules
        }
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
