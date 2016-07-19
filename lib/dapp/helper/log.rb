module Dapp
  module Helper
    # Log
    module Log
      def log(message = '', indent: false, desc: nil)
        return unless defined?(cli_options) and !cli_options[:log_quiet]
        unless desc.nil?
          (desc[:data] ||= {})[:msg] = message
          message = t(desc: desc)
        end
        if indent
          log_with_indent(message)
        else
          puts message.to_s.lines.map { |line| ' ' * 2 * cli_options[:log_indent].to_i + line }.join
        end
      end

      def log_with_indent(message = '')
        with_log_indent do
          log(message)
        end
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
    end # Log
  end # Helper
end # Dapp
