module Dapp
  module Helper
    # Streaming
    module Streaming
      class Stream
        def buffer
          @buffer ||= []
        end

        def <<(string)
          buffer << string
        end

        def inspect
          buffer.join
        end
      end

      module Proxy
        class Base
          include Paint

          def initialize(*streams)
            @streams = streams
          end

          def <<(str)
            @streams.each { |s| s << format_string(str) }
          end

          def format_string(str)
            str
          end
        end

        class Error < Base
          def format_string(str)
            Paint.paint_string(str.strip, :warning)
          end
        end
      end
    end
  end # Helper
end # Dapp
