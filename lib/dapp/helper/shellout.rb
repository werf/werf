module Dapp
  module Helper
    # Shellout
    module Shellout
      include Streaming

      def shellout(*args, log_verbose: false, **kwargs)
        do_shellout = proc do
          log_verbose = (log_verbose && cli_options[:log_verbose]) if defined? cli_options
          kwargs[:live_stream] ||= STDOUT if log_verbose
          ::Mixlib::ShellOut.new(*args, timeout: 3600, **kwargs).run_command
        end

        if defined? ::Bundler
          ::Bundler.with_clean_env { do_shellout.call }
        else
          do_shellout.call
        end
      end

      def shellout!(*args, log_verbose: false, **kwargs)
        stream = Stream.new
        if log_verbose
          kwargs[:live_stream] = Proxy::Base.new(stream, STDOUT)
        else
          kwargs[:live_stdout] ||= stream
        end
        kwargs[:live_stderr] ||= Proxy::Error.new(stream)
        shellout(*args, **kwargs).tap(&:error!)
      rescue ::Mixlib::ShellOut::ShellCommandFailed => e
        raise Error::Shellout, code: Trivia.class_to_lowercase(e.class),
                               data: { stream: stream.inspect }
      end

      def self.included(base)
        base.extend(self)
      end
    end
  end # Helper
end # Dapp
