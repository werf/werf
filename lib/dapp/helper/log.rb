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

      def log(message = '', desc: nil, style: nil, indent: false, ignore_indent: false, new_line: true)
        return unless defined?(cli_options) and !cli_options[:log_quiet]
        unless desc.nil?
          (desc[:data] ||= {})[:msg] = message
          message = t(desc: desc)
        end
        formatted_message = style ? Paint[message, *style(style)] : message
        if indent
          log_with_indent(formatted_message)
        else
          print formatted_message.to_s.lines.map { |line| ignore_indent ? line : (log_indent + line) }.join
          print "\n" if new_line
        end
      end

      def log_with_indent(message = '', **kvargs)
        with_log_indent do
          log(message, **kvargs)
        end
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

      def with_log_indent(with = true)
        log_indent_next if with
        yield
        log_indent_prev if with
      end

      def style(name)
        public_send("log_#{name}_format")
      end

      def log_step_format
        [:yellow, :bold]
      end

      def log_info_format
        [:blue]
      end

      def log_failed_format
        [:red, :bold]
      end

      def log_success_format
        [:green, :bold]
      end

      def log_secondary_format
        [:white, :bold]
      end

      def self.error_colorize(error_msg)
        Paint[error_msg, :red, :bold]
      end
    end # Log
  end # Helper
end # Dapp
