module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module FlushRepo
            def stages_flush_repo
              ruby2go_cleanup_command(:flush, ruby2go_cleanup_stages_flush_repo_options)
            end

            def ruby2go_cleanup_stages_flush_repo_options
              {
                common_repo_options: {
                  repository: option_repo,
                  dimg_names: dimgs_names,
                  dry_run: dry_run?
                },
                mode: {
                  with_dimgs: false,
                  with_stages: true,
                  only_repo: true,
                },
              }.tap do |json|
                break JSON.dump(json)
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
