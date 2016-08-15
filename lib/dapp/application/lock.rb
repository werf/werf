module Dapp
  # Application
  class Application
    # Lock
    module Lock
      def lock(name, *args, default_timeout: 60, &blk)
        timeout = @cli_options[:lock_timeout] || default_timeout
        ::Dapp::Lock::File.new(self.lock_path, name, timeout: timeout).synchronize(*args, &blk)
      end
    end # Lock
  end # Application
end # Dapp
