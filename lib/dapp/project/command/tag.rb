module Dapp
  # Project
  class Project
    # Command
    module Command
      # Tag
      module Tag
        def tag(tag)
          one_dimg!
          raise Error::Project, code: :tag_command_incorrect_tag, data: { name: tag } unless Image::Docker.image_name?(tag)
          Dimg.new(config: build_configs.first, project: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
            app.tag!(tag)
          end
        end
      end
    end
  end # Project
end # Dapp
