module Dapp
  module Deployment
    module Config
      module Directive
        class Expose < Base
          attr_reader :_port
          attr_reader :_type

          def initialize(dapp:)
            @_port = []
            @_type = 'ClusterIP'
            super
          end

          def cluster_ip
            sub_directive_eval { @_type = 'ClusterIP' }
          end

          def load_balancer
            sub_directive_eval { @_type = 'LoadBalancer' }
          end

          def node_port
            sub_directive_eval { @_type = 'NodePort' }
          end

          def port(number, &blk)
            sub_directive_eval { @_port << Port.new(number, dapp: dapp, &blk) }
          end

          class Port < Base
            attr_reader :_number, :_target, :_protocol

            def initialize(number, dapp:, &blk)
              self._number = number
              @_protocol = 'TCP'
              super(dapp: dapp, &blk)
            end

            def target(number)
              @_target = define_number(number, :unsupported_target_number)
            end

            def tcp
              @_protocol = 'TCP'
            end

            def udp
              @_protocol = 'UDP'
            end

            def _number=(number)
              @_number = define_number(number, :unsupported_port_number)
            end

            protected

            def define_number(number, code)
              number.to_i.tap do |n|
                raise Error::Config, code: code, data: { number: number } unless (0..65536).cover?(n)
              end
            end
          end
        end
      end
    end
  end
end
