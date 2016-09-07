module Dapp
  # Project
  class Project
    # Command
    module Command
      # StageImage
      module StageImage
        def stage_image
          raise Error::Project, code: :command_unexpected_apps_number unless build_configs.one?
          puts Application.new(config: build_configs.first, project: self, ignore_git_fetch: true).stage_image_name(cli_options[:stage])
        end
      end
    end
  end # Project
end # Dapp
