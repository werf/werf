module Dapp
  module Helper
    # Streaming
    module Streaming
      # Stream
      class Stream
        def buffer
          @buffer ||= []
        end

        def <<(string)
          buffer << string
        end

        def inspect
          buffer.join("\n")
        end
      end

      # Proxy
      module Proxy
        # Base
        class Base
          include Helper::Log

          def initialize(*streams, with_time: false)
            @streams = streams
            @with_time = with_time
          end

          def <<(str)
            str = format_string(str)
            @streams.each { |s| s << "#{str}#{"\n" if s.is_a?(IO)}" }
          end

          def format_string(str)
            str.lines.map { |l| "#{log_time if @with_time}#{l.strip}" }.join
          end
        end

        # Error
        class Error < Base
          def format_string(str)
            Paint.paint_string(super, :warning)
          end
        end
      end
    end
  end # Helper
end # Dapp
