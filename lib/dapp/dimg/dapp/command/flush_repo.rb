module Dapp
  module Dimg
    module Dapp
      module Command
        module FlushRepo
          def flush_repo
            ruby2go_cleanup_command(:flush, ruby2go_cleanup_flush_repo_options_dump)
          end

          def ruby2go_cleanup_flush_repo_options_dump
            ruby2go_cleanup_common_repo_options.merge(
              mode: {
                with_dimgs: true,
                with_stages: with_stages?,
                only_repo: true,
              }
            ).tap do |json|
              break JSON.dump(json)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
