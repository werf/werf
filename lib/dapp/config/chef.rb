module Dapp
  module Config
    class Chef
      attr_reader :_module

      def initialize
        @_module = []
      end

      def module(*args)
        @_module.push(*args.flatten)
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
