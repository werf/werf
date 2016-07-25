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

      def paint_style(name)
        FORMAT[name].tap do |format|
          fail if format.nil?
        end
      end

      def paint_string(object, style_name)
        ::Paint[::Paint.unpaint(object.to_s), *paint_style(style_name)]
      end
    end # Paint
  end # Helper
end # Dapp

Dapp::Helper::Paint.extend Dapp::Helper::Paint
