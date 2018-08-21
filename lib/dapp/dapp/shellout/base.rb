module Dapp
  class Dapp
    module Shellout
      module Base
        include Streaming

        def shellout(*args, **kwargs)
          env = nil

          Base.default_env_keys.each do |env_key|
            env_key = env_key.to_s.upcase
            env ||= {}
            env[env_key] = ENV[env_key]
          end

          param_env = (kwargs.delete(:env) || kwargs.delete(:environment))
          param_env = param_env.map { |key, value| [key.to_s.upcase, value.to_s] }.to_h if param_env

          env = (env || {}).merge(param_env) if param_env
          kwargs[:env] = env if env

          do_shellout = -> { ::Mixlib::ShellOut.new(*args, timeout: 86400, **kwargs).run_command }
          if defined? ::Bundler
            ::Bundler.with_clean_env { do_shellout.call }
          else
            do_shellout.call
          end
        end

        def shellout!(*args, **kwargs)
          if instance_of? Dapp
            default_kwarg = proc { |key, value| kwargs[key] = value unless kwargs.key?(key) }
            default_kwarg.call(:quiet, log_quiet?)
            default_kwarg.call(:time, log_time?)
          end
          _shellout_with_logging!(*args, **kwargs)
        end

        def shellout_pack(command)
          "eval $(echo #{Base64.strict_encode64(command)} | #{base64_bin} --decode)"
        end

        class << self
          def included(base)
            base.extend(self)
          end

          def default_env_keys
            @default_env_keys ||= []
          end
        end # << self

        def shellout_cmd_should_succeed!(cmd)
          return cmd.tap(&:error!)
        rescue ::Mixlib::ShellOut::ShellCommandFailed => e
          raise Error::Shellout, code: Helper::Trivia.class_to_lowercase(e.class), data: { stream: e.message }
        end

        protected

        def _shellout_with_logging!(*args, verbose: false, quiet: true, time: false, raise_on_error: true, **kwargs)
          stream = Stream.new
          if verbose && !quiet
            kwargs[:live_stream] = Proxy::Base.new(stream, STDOUT, with_time: time)
            kwargs[:live_stderr] = Proxy::Error.new(stream, STDERR, with_time: time)
          else
            kwargs[:live_stdout] = Proxy::Base.new(stream, with_time: time)
            kwargs[:live_stderr] = Proxy::Error.new(stream, with_time: time)
          end

          shellout(*args, **kwargs).tap do |res|
            res.error! if raise_on_error
          end
        rescue ::Mixlib::ShellOut::ShellCommandFailed => e
          raise Error::Shellout, code: Helper::Trivia.class_to_lowercase(e.class), data: { stream: stream.show }
        end
      end
    end
  end # Helper
end # Dapp
