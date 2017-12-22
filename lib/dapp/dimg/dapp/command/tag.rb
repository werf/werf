module Dapp
  module Dimg
    module Dapp
      module Command
        module Tag
          def tag
            dimg_import_export_base do |dimg|
              dimg.tag!(option_repo, format: push_format(dimg.config._name))
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
