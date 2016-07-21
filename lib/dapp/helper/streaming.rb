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
          buffer.join
        end
      end

      # Proxy
      module Proxy
        # Base
        class Base
          def initialize(*streams)
            @streams = streams
          end

          def <<(str)
            @streams.each { |s| s << format_string(str) }
          end

          def format_string(str)
            str.strip
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
