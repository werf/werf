module Dapp
  module Config
    # Chef
    class Chef
      attr_reader :_module

      def initialize
        @_module = [] # FIXME -> modules
      end

      def module(*args)
        @_module.concat(args)
      end

      def to_h
        {
          module: _module
        }
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
