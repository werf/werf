module Dapp
  # Application
  class Application
    # Lock
    module Lock
      def lock(name, *args, default_timeout: 60, &blk)
        timeout = @cli_options[:lock_timeout] || default_timeout

        ::Dapp::Lock::File.new(
          self.lock_path, name,
          timeout: timeout,
          on_wait: ->(&blk) {
            log_secondary_process(
              self.t(code: 'process.waiting_resouce_lock', data: { name: name }),
              short: true,
              &blk
            )
          }
        ).synchronize(*args, &blk)
      end
    end # Lock
  end # Application
end # Dapp
