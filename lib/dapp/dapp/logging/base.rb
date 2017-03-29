module Dapp
  class Dapp
    module Logging
      module Base
        def log_quiet?
          cli_options[:log_quiet]
        end

        def log_time?
          cli_options[:log_time]
        end

        def log_verbose?
          cli_options[:log_verbose]
        end

        def ignore_config_warning?
          cli_options[:ignore_config_warning]
        end

        def introspect_error?
          cli_options[:introspect_error]
        end

        def introspect_before_error?
          cli_options[:introspect_before_error]
        end

        def dry_run?
          cli_options[:dry_run]
        end

        def dev_mode?
          cli_options[:dev].nil? ? config._dev_mode : cli_options[:dev]
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
          with_log_indent do
            yield
          end
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
          "#{DateTime.now.strftime('%Y-%m-%dT%T%z')} "
        end

        def log_format_string(str, time: true, indent: true, style: nil)
          str.to_s.lines.map do |line|
            line = paint_string(line, style) if style
            "#{log_time if time && log_time?}#{indent ? (log_indent + line) : line}"
          end.join
        end

        def log_with_indent(message = '', **kwargs)
          with_log_indent do
            log(message, **kwargs)
          end
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

        class << self
          def included(base)
            base.extend(self)
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
