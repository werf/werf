module Dapp
  # Project
  class Project
    # Command
    module Command
      # Tag
      module Tag
        def tag(tag)
          raise Error::Project, code: :tag_command_unexpected_dimgs_number unless build_configs.one?
          raise Error::Project, code: :tag_command_incorrect_tag, data: { name: tag } unless Image::Docker.image_name?(tag)
          Application.new(config: build_configs.first, project: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
            app.tag!(tag)
          end
        end
      end
    end
  end # Project
end # Dapp
