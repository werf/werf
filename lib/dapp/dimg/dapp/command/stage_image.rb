module Dapp
  module Dimg
    module Dapp
      module Command
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
