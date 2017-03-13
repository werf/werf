module Dapp
  class Dapp
    module Logging
      module Paint
        FORMAT = {
          step: [:yellow, :bold],
          info: [:blue],
          success: [:green, :bold],
          warning: [:red, :bold],
          secondary: [:white, :bold],
          default: [:white]
        }.freeze

        def self.initialize(mode)
          ::Paint.mode = case mode
                         when 'auto' then (ENV['TRAVIS'] || ENV['GITLAB_CI'] || STDOUT.tty?) ? 8 : 0
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
          ::Paint[unpaint(object.to_s), *paint_style(style_name)]
        end

        def unpaint(str)
          ::Paint.unpaint(str)
        end

        class << self
          def included(base)
            base.extend(self)
          end
        end
      end # Paint
    end
  end # Dapp
end # Dapp
