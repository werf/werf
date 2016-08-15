module Dapp
  # Application
  class Application
    # Lock
    module Lock
      def lock(name, *args, default_timeout: 60, &blk)
        ::Dapp::Lock::File.new(self.lock_path, name, timeout: default_timeout).synchronize(*args, &blk)
      end
    end # Lock
  end # Application
end # Dapp
