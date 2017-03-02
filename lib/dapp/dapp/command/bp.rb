module Dapp
  # Dapp
  class Dapp
    # Command
    module Command
      # Build
      module Bp
        def bp(repo)
          bp_step(:build)
          bp_step(:push, repo)
          bp_step(:stages_cleanup_by_repo, repo)
          bp_step(:cleanup)
        end

        def bp_step(step, *args)
          log_step_with_indent(step) { send(step, *args) }
        end
      end
    end
  end # Dapp
end # Dapp
