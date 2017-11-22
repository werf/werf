module Dapp
  class Dapp
    module Lock
      def lock_path(type)
        case type
        when :dapp
          build_path.join('locks')
        when :global
          Pathname.new("/tmp/dapp.global.locks").tap do |p|
            FileUtils.mkdir_p p.to_s
          end
        else
          raise
        end
      end

      def _lock(name, type)
        @_locks ||= {}
        @_locks[type] ||= {}
        @_locks[type][name] ||= ::Dapp::Dimg::Lock::File.new(lock_path(type), name)
      end

      def lock(name, *_args, default_timeout: 300, type: :dapp, **kwargs, &blk)
        if dry_run?
          yield if block_given?
        else
          timeout = options[:lock_timeout] || default_timeout
          _lock(name, type).synchronize(
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
