module Dapp
  module Helper
    # Log
    module Log
      def log_info(message, *args)
        log(message, *args, style: :info)
      end

      def log_step(message, *args)
        log(message, *args, style: :step)
      end

      def log_secondary(message, *args)
        log(message, *args, style: :secondary)
      end

      def log(message = '', desc: nil, inline: false, **kwargs)
        return unless defined?(cli_options) && !cli_options[:log_quiet]
        unless desc.nil?
          (desc[:data] ||= {})[:msg] = message
          message = t(desc: desc)
        end
        print "#{log_format_string(message, **kwargs)}#{"\n" unless inline}"
      end

      def log_time
        "#{DateTime.now.strftime('%Y-%m-%dT%T%z')} "
      end

      def log_format_string(str, time: true, indent: true, style: nil)
        str.to_s.lines.map do |line|
          line = paint_string(line, style) if style
          "#{log_time if time && cli_options[:log_time]}#{indent ? (log_indent + line) : line}"
        end.join
      end

      def log_with_indent(message = '', **kvargs)
        with_log_indent do
          log(message, **kvargs)
        end
      end

      def with_log_indent(with = true)
        log_indent_next if with
        yield
        log_indent_prev if with
      end

      def log_indent
        ' ' * 2 * cli_options[:log_indent].to_i
      end

      def log_indent_next
        return unless defined? cli_options
        cli_options[:log_indent] += 1
      end

      def log_indent_prev
        return unless defined? cli_options
        if cli_options[:log_indent] <= 0
          cli_options[:log_indent] = 0
        else
          cli_options[:log_indent] -= 1
        end
      end

      def self.included(base)
        base.include(Paint)
      end
    end # Log
  end # Helper
end # Dapp
