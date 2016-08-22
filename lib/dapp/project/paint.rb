module Dapp
  # Project
  class Project
    # Paint
    module Paint
      def paint_initialize
        ::Paint.mode = case cli_options[:log_color]
          when 'auto' then STDOUT.tty? ? 8 : 0
          when 'on'   then 8
          when 'off'  then 0
          else raise
        end
      end
    end # Paint
  end # Project
end # Dapp
