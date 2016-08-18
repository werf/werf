module Dapp
  # Application
  class Application
    # Lock
    module Lock
      def lock(name, *args, default_timeout: 60, **kwargs, &blk)
        timeout = @cli_options[:lock_timeout] || default_timeout

        ::Dapp::Lock::File.new(
          lock_path, name,
          timeout: timeout,
          on_wait: ->(&blk) {
            log_secondary_process(
              self.t(code: 'process.waiting_resouce_lock', data: { basename: home_path.basename,
                                                                   name: name }),
              short: true,
              &blk
            )
          }
        ).synchronize(*args, **kwargs, &blk)
      rescue Dapp::Lock::Error::Timeout => e
        raise Dapp::Lock::Error::Timeout, e.net_status.tap { |err| err[:data][:basename] = home_path.basename }
      end
    end # Lock
  end # Application
end # Dapp
