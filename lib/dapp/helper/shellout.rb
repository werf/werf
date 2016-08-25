module Dapp
  module Helper
    # Shellout
    module Shellout
      include Streaming

      def shellout(*args, log_verbose: false, **kwargs)
        log_verbose = (log_verbose && cli_options[:log_verbose]) if defined? cli_options
        kwargs[:live_stream] ||= STDOUT if log_verbose

        env = nil

        Shellout.default_env_keys.each do |env_key|
          env_key = env_key.to_s.upcase
          env ||= {}
          env[env_key] = ENV[env_key]
        end

        param_env = (kwargs.delete(:env) || kwargs.delete(:environment))
        param_env = param_env.map { |key, value| [key.to_s.upcase, value.to_s] }.to_h if param_env

        env = (env || {}).merge(param_env) if param_env
        kwargs[:env] = env if env

        do_shellout = -> { ::Mixlib::ShellOut.new(*args, timeout: 3600, **kwargs).run_command }
        if defined? ::Bundler
          ::Bundler.with_clean_env { do_shellout.call }
        else
          do_shellout.call
        end
      end

      def shellout!(*args, log_verbose: false, log_time: false, **kwargs)
        stream = Stream.new
        log_time = defined?(cli_options) ? cli_options[:log_time] : log_time
        if log_verbose
          kwargs[:live_stream] = Proxy::Base.new(stream, STDOUT, with_time: log_time)
        else
          kwargs[:live_stdout] = Proxy::Base.new(stream, with_time: log_time)
        end
        kwargs[:live_stderr] = Proxy::Error.new(stream, with_time: log_time)
        shellout(*args, **kwargs).tap(&:error!)
      rescue ::Mixlib::ShellOut::ShellCommandFailed => e
        raise Error::Shellout, code: Trivia.class_to_lowercase(e.class),
                               data: { stream: stream.show }
      end

      def shellout_pack(command)
        "eval $(echo #{Base64.strict_encode64(command)} | base64 --decode)"
      end

      class << self
        def included(base)
          base.extend(self)
        end

        def default_env_keys
          @default_env_keys ||= []
        end
      end # << self
    end
  end # Helper
end # Dapp
