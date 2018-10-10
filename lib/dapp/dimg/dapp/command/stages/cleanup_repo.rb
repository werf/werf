module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module CleanupRepo
            def stages_cleanup_repo
              lock_repo(repo = option_repo) do
                raise ::Dapp::Error::Command, code: :stages_cleanup_required_option unless stages_cleanup_option?

                ruby2go_cleanup_command(:sync, cleanup_repo_proper_cache_version_options) if proper_cache_version?
                repo_dimgstages_cleanup if proper_repo_cache?
              end
            end

            def repo_dimgstages_cleanup
              ruby2go_cleanup_command(:sync, cleanup_repo_proper_repo_cache_options)
            end

            def cleanup_repo_proper_cache_version_options
              ruby2go_cleanup_sync_common_repo_options.merge({ mode: { sync_repo: true, only_cache_version: true } }).merge(ruby2go_cleanup_sync_cache_version_option).tap do |data|
                break JSON.dump(data)
              end
            end

            def cleanup_repo_proper_repo_cache_options
              ruby2go_cleanup_sync_common_repo_options.merge({ sync_repo: true }).tap do |data|
                break JSON.dump(data)
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
