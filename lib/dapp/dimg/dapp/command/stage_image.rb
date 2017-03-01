module Dapp
  module Dimg
    # Dapp
    class Dapp
      # Command
      module Command
        # StageImage
        module StageImage
          def stage_image
            one_dimg!
            puts Dimg.new(config: build_configs.first, dapp: self, ignore_git_fetch: true).stage_image_name(cli_options[:stage])
          end
        end
      end
    end
  end # Dimg
end # Dapp
