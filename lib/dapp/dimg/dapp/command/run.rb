module Dapp
  module Dimg
    module Dapp
      module Command
        module Run
          def run(stage_name, docker_options, command)
            one_dimg!
            dimg(config: build_configs.first, should_be_built: stage_name.nil?, ignore_signature_auto_calculation: true)
              .run_stage(stage_name, docker_options, command)
          end
        end
      end
    end
  end # Dimg
end # Dapp
