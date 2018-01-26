module Dapp
  class Dapp
    module Lock
      def locks_dir
        File.join(self.class.home_dir, "locks")
      end

      def _lock(name)
        @_locks ||= {}
        @_locks[name] ||= ::Dapp::Dimg::Lock::File.new(locks_dir, name)
      end

      # default_timeout 24 hours
      def lock(name, *_args, default_timeout: 86400, **kwargs, &blk)
        if dry_run?
          yield if block_given?
        else
          timeout = options[:lock_timeout] || default_timeout
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
