module Dapp
  class Dapp
    module Logging
      module Base
        class << self
          def included(base)
            base.send(:extend, ClassMethods)
          end
        end

        module ClassMethods
          def log_time
            "#{DateTime.now.strftime('%Y-%m-%dT%T%z')} "
          end
        end

        def log_quiet?
          option_quiet
        end

        def log_time?
          option_time
        end

        def log_verbose?
          option_verbose
        end

        def ignore_config_warning?
          !!options[:ignore_config_warning]
        end

        def introspect_error?
          !!options[:introspect_error]
        end

        def introspect_before_error?
          !!options[:introspect_before_error]
        end

        def dry_run?
          option_dry_run
        end

        def dev_mode?
          option_dev
        end

        def log_info(*args, **kwargs)
          kwargs[:style] = :info
          log(*args, **kwargs)
        end

        def log_dimg_name_with_indent(dimg, &blk)
          return yield if dimg._name.nil?
          log_step_with_indent(dimg._name, &blk)
        end

        def log_step_with_indent(step)
          log_step(step)
          with_log_indent { yield }
        end

        def log_step(*args, **kwargs)
          kwargs[:style] = :step
          log(*args, **kwargs)
        end

        def log_secondary(*args, **kwargs)
          kwargs[:style] = :secondary
          log(*args, **kwargs)
        end

        def log_warning(*args, **kwargs)
          kwargs[:style] = :warning
          kwargs[:stream] ||= $stderr
          if args.empty?
            kwargs[:desc] ||= {}
            kwargs[:desc][:context] ||= :warning
          end
          log(*args, **kwargs)
        end

        def log_config_warning(*args, **kwargs)
          return if ignore_config_warning?
          log_warning(*args, **kwargs)
        end

        def log(message = '', desc: nil, inline: false, stream: $stdout, **kwargs)
          return if log_quiet?
          unless desc.nil?
            (desc[:data] ||= {})[:msg] = message
            message = t(**desc)
          end
          stream.print "#{log_format_string(message, **kwargs)}#{"\n" unless inline}"
        end

        def log_time
          self.class.log_time
        end

        def log_format_string(str, time: true, indent: true, style: nil)
          str.to_s.lines.map do |line|
            line = paint_string(line, style) if style
            "#{log_time if time && log_time?}#{indent ? (log_indent + line) : line}"
          end.join
        end

        def with_log_indent(with = true)
          log_indent_next if with
          yield
        ensure
          log_indent_prev if with
        end

        def log_indent
          ' ' * 2 * log_indent_size
        end

        def log_indent_next
          self.log_indent_size += 1
        end

        def log_indent_prev
          if self.log_indent_size <= 0
            self.log_indent_size = 0
          else
            self.log_indent_size -= 1
          end
        end

        def service_stream
          @@service_stream ||= begin
            fd = IO.sysopen("/tmp/dapp-service.log", "a")

            stream = IO.new(fd)
            stream.sync = true
            stream
          end
        end

        protected

        attr_writer :log_indent_size

        def log_indent_size
          @log_indent_size ||= 0
        end
      end
    end # Logging
  end # Dapp
end # Dapp
