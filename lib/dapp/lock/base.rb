module Dapp
  # Lock
  module Lock
    # Base
    class Base
      attr_reader :name
      attr_reader :timeout

      def initialize(name, timeout: 60)
        @name = name
        @timeout = timeout
      end

      def lock(shared: false)
        raise
      end

      def unlock
        raise
      end

      def synchronize(*args)
        lock(*args)
        begin
          yield if block_given?
        ensure
          unlock
        end
      end
    end # Base
  end # Lock
end # Dapp
