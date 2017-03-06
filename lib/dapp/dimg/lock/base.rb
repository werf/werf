module Dapp
  module Dimg
    module Lock
      class Base
        attr_reader :name

        def initialize(name)
          @name = name
          @active_locks = 0
        end

        def lock(timeout: 60, on_wait: nil, readonly: false)
          _do_lock(timeout, on_wait, readonly) unless @active_locks > 0
          @active_locks += 1
        end

        def unlock
          @active_locks -= 1
          _do_unlock if @active_locks.zero?
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

        def _do_lock(_timeout, _on_wait, _readonly)
          raise
        end

        def _do_unlock
          raise
        end

        def _waiting(timeout, on_wait, &blk)
          if on_wait
            on_wait.call { ::Timeout.timeout(timeout, &blk) }
          else
            ::Timeout.timeout(timeout, &blk)
          end
        end
      end # Base
    end # Lock
  end
end # Dapp
