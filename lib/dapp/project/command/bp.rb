module Dapp
  # Project
  class Project
    # Command
    module Command
      # Build
      module Bp
        def bp(repo)
          bp_step(:build)
          bp_step(:push, repo)
          bp_step(:stages_cleanup_local, repo)
          bp_step(:cleanup)
        end

        def bp_step(step, *args)
          log_step_with_indent(step) { public_send(step, *args) }
        end
      end
    end
  end # Project
end # Dapp
