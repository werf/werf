module Dapp
  module Helper
    # Paint
    module Paint
      FORMAT = {
        step: [:yellow, :bold],
        info: [:blue],
        success: [:green, :bold],
        warning: [:red, :bold],
        secondary: [:white, :bold],
        default: [:white]
      }.freeze

      def self.initialize
        return unless defined?(cli_options)
        Paint.mode = case cli_options[:log_color]
                     when 'auto' then STDOUT.tty? ? 8 : 0
                     when 'on'   then 8
                     when 'off'  then 0
                     else raise
                     end
      end

      def paint_style(name)
        FORMAT[name].tap do |format|
          raise if format.nil?
        end
      end

      def paint_string(object, style_name)
        ::Paint[::Paint.unpaint(object.to_s), *paint_style(style_name)]
      end
    end # Paint
  end # Helper
end # Dapp

Dapp::Helper::Paint.extend Dapp::Helper::Paint
