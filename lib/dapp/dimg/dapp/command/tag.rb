module Dapp
  module Dimg
    module Dapp
      module Command
        module Tag
          def tag(tag_or_shortcut)
            one_dimg!
            tag = resolve_tag(tag_or_shortcut)
            validate_image_name!(tag)

            Dimg.new(config: build_configs.first, dapp: self, ignore_git_fetch: true, should_be_built: true).tap do |app|
              app.tag!(tag)
            end
          end

          def resolve_tag(tag_or_shortcut)
            if (repo = shortcuts[tag_or_shortcut])
              dimg_name = dimg_name!
              format(push_format(dimg_name), repo: repo, dimg_name: dimg_name, tag: 'latest')
            else
              tag_or_shortcut
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
