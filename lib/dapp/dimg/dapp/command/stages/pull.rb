module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module Pull
            def stages_pull
              dimg_import_export_base(should_be_built: false) do |dimg|
                dimg.import_stages!(option_repo, format: dimgstage_push_format)
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
