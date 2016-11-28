module Dapp
  # Project
  class Project
    # Command
    module Command
      # StageImage
      module StageImage
        def stage_image
          one_dimg!
          puts Dimg.new(config: build_configs.first, project: self, ignore_git_fetch: true).stage_image_name(cli_options[:stage])
        end
      end
    end
  end # Project
end # Dapp
