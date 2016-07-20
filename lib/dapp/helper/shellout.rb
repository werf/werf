module Dapp
  module Helper
    # Shellout
    module Shellout
      def shellout(*args, log_verbose: false, **kwargs)
        do_shellout = proc do
          log_verbose = (log_verbose && cli_options[:log_verbose]) if defined? cli_options
          kwargs[:live_stream] = STDOUT if log_verbose
          Mixlib::ShellOut.new(*args, timeout: 3600, **kwargs).run_command
        end

        if defined? ::Bundler
          ::Bundler.with_clean_env { do_shellout.call }
        else
          do_shellout.call
        end
      end

      def shellout!(*args, **kwargs)
        shellout(*args, **kwargs).tap(&:error!)
      end

      def self.included(base)
        base.extend(self)
      end
    end
  end # Helper
end # Dapp
