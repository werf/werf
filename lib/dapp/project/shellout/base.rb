module Dapp
  class Project
    # Shellout
    module Shellout
      # Base
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

          do_shellout = -> { ::Mixlib::ShellOut.new(*args, timeout: 3600, **kwargs).run_command }
          if defined? ::Bundler
            ::Bundler.with_clean_env { do_shellout.call }
          else
            do_shellout.call
          end
        end

        def shellout!(*args, **kwargs)
          shellout_with_logging(**kwargs) do |options|
            shellout(*args, **options).tap(&:error!)
          end
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

        protected

        def shellout_with_logging(log_verbose: false, **kwargs)
          return yield(**kwargs) unless instance_of? Project

          begin
            stream = Stream.new
            if log_verbose && log_verbose?
              kwargs[:live_stream] = Proxy::Base.new(stream, STDOUT, with_time: log_time?)
            else
              kwargs[:live_stdout] = Proxy::Base.new(stream, with_time: log_time?)
            end
            kwargs[:live_stderr] = Proxy::Error.new(stream, with_time: log_time?)

            yield(**kwargs)
          rescue ::Mixlib::ShellOut::ShellCommandFailed => e
            raise Error::Shellout, code: class_to_lowercase(e.class), data: { stream: stream.show, backtrace: e.backtrace.join("\n") }
          end
        end
      end
    end
  end # Helper
end # Dapp
