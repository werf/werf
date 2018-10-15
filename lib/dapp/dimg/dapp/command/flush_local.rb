module Dapp
  module Dimg
    module Dapp
      module Command
        module FlushLocal
          def flush_local
            ruby2go_cleanup_command(:flush, ruby2go_cleanup_flush_local_options_dump)
          end

          def ruby2go_cleanup_flush_local_options_dump
            ruby2go_cleanup_common_project_options(force: true).merge(
              mode: {
                with_dimgs: true,
                with_stages: with_stages?,
                only_repo: false,
              },
            ).tap do |json|
              break JSON.dump(json)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
