module Dapp
  module Deployment
    module Config
      module Directive
        class Expose < Base
          attr_reader :_port, :_cluster_ip

          def initialize(dapp:)
            @_port = []
            super
          end

          def cluster_ip
            sub_directive_eval { @_cluster_ip = true }
          end

          def port(*args, &blk)
            sub_directive_eval { @_port << Port.new(*args, dapp: dapp, &blk) }
          end

          class Port < Base
            attr_reader :_list, :_protocol

            def initialize(*args, dapp:, &blk)
              self._list = args
              @_protocol = :TCP
              super(dapp: dapp, &blk)
            end

            def tcp
              @_protocol = :TCP
            end

            def udp
              @_protocol = :UDP
            end

            def _list=(ports)
              @_list = begin
                ports.map do |port|
                  port.to_i.tap do |p|
                    raise Error::Config, code: :unsupported_port_number, data: { port: port } unless (0..65536).cover?(p)
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
