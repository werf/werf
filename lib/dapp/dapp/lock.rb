module Dapp
  # Dapp
  class Dapp
    # Lock
    module Lock
      def lock_path
        build_path.join('locks')
      end

      def _lock(name)
        @_locks ||= {}
        @_locks[name] ||= ::Dapp::Dimg::Lock::File.new(lock_path, name)
      end

      def lock(name, *_args, default_timeout: 300, **kwargs, &blk)
        if dry_run?
          yield if block_given?
        else
          timeout = cli_options[:lock_timeout] || default_timeout
          _lock(name).synchronize(
            timeout: timeout,
            on_wait: proc do |&do_wait|
              log_secondary_process(
                t(code: 'process.waiting_resouce_lock', data: { name: name }),
                short: true,
                &do_wait
              )
            end, **kwargs, &blk
          )
        end
      end
    end # Lock
  end # Dapp
end # Dapp
