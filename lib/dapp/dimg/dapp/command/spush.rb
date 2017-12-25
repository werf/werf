module Dapp
  module Dimg
    module Dapp
      module Command
        module Spush
          def spush
            one_dimg!
            dimg_import_export_base do |dimg|
              dimg.export!(option_repo, format: spush_format)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
