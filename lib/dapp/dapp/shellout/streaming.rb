module Dapp
  class Dapp
    module Shellout
      module Streaming
        class Stream
          def buffer
            @buffer ||= []
          end

          def <<(string)
            buffer << string
          end

          def show
            buffer.join.strip
          end
        end

        module Proxy
          class Base
            def initialize(*streams, with_time: false)
              @streams = streams
              @with_time = with_time
            end

            def <<(str)
              @streams.each { |s| s << format_string(str) }
            end

            def format_string(str)
              str.lines.map { |l| "#{Dapp.log_time if @with_time}#{l.chomp}\n" }.join
            end
          end

          class Error < Base
            def format_string(str)
              "#{Dapp.paint_string(super.chomp, :warning)}\n"
            end
          end
        end
      end
    end
  end # Helper
end # Dapp
