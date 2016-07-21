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

      # rubocop:disable Metrics/CyclomaticComplexity
      def log(message = '', desc: nil, style: nil, indent: true, inline: false)
        return unless defined?(cli_options) && !cli_options[:log_quiet]
        unless desc.nil?
          (desc[:data] ||= {})[:msg] = message
          message = t(desc: desc)
        end
        formatted_message = begin
          message = paint_string(message, style) if style
          message.to_s.lines.map { |line| indent ? (log_indent + line) : line }.join
        end
        print "#{formatted_message}#{"\n" unless inline}"
      end
      # rubocop:enable Metrics/CyclomaticComplexity

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
