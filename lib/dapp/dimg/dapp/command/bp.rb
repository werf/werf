module Dapp
  module Dimg
    module Dapp
      module Command
        module Bp
          def bp
            bp_step(:build)
            bp_step(:push)
          end

          def bp_step(step, *args)
            log_step_with_indent(step) { send(step, *args) }
          end
        end
      end
    end
  end # Dimg
end # Dapp
