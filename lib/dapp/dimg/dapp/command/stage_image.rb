module Dapp
  module Dimg
    module Dapp
      module Command
        module StageImage
          def stage_image
            one_dimg!
            puts dimg(config: build_configs.first, ignore_git_fetch: true).stage_image_name(options[:stage])
          end
        end
      end
    end
  end # Dimg
end # Dapp
