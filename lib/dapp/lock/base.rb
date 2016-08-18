module Dapp
  # Lock
  module Lock
    # Base
    class Base
      attr_reader :name
      attr_reader :on_wait
      attr_reader :timeout

      def initialize(name, timeout: 60, on_wait: nil)
        @name = name
        @on_wait = on_wait
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

      protected

      def _waiting(&blk)
        if @on_wait
          @on_wait.call { ::Timeout.timeout(timeout, &blk) }
        else
          ::Timeout.timeout(timeout, &blk)
        end
      end
    end # Base
  end # Lock
end # Dapp
