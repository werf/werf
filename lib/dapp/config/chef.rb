module Dapp
  module Config
    class Chef < Base
      attr_reader :_module

      def initialize
        @_module = []
        super
      end

      def module(*args)
        @_module.push(*args.flatten)
      end

      def to_h
        {
          module: _module
        }
      end
    end
  end
end
