module Dapp
  # Project
  class Project
    # Lock
    module Lock
      def lock_path
        build_path.join('locks')
      end

      def lock(name, *args, default_timeout: 60, **kwargs, &blk)
        timeout = cli_options[:lock_timeout] || default_timeout

        ::Dapp::Lock::File.new(
          lock_path, name,
          timeout: timeout,
          on_wait: ->(&blk) {
            log_secondary_process(
              self.t(code: 'process.waiting_resouce_lock', data: { name: name }),
              short: true,
              &blk
            )
          }
        ).synchronize(*args, **kwargs, &blk)
      end
    end # Lock
  end # Project
end # Dapp
