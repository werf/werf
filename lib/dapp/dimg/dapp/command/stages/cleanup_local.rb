module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module CleanupLocal
            def stages_cleanup_local
              raise ::Dapp::Error::Command, code: :stages_cleanup_required_option unless stages_cleanup_option?

              if proper_cache_version?
                ruby2go_cleanup_command(:sync, cleanup_local_proper_cache_version_options)
              end

              if proper_repo_cache?
                ruby2go_cleanup_command(:sync, cleanup_local_proper_repo_cache_options)
              end
            end

            def cleanup_local_proper_cache_version_options
              ruby2go_cleanup_sync_common_project_options.merge(ruby2go_cleanup_sync_cache_version_option).tap do |data|
                break JSON.dump(data)
              end
            end

            def cleanup_local_proper_repo_cache_options
              ruby2go_cleanup_sync_common_project_options.merge(ruby2go_cleanup_sync_common_repo_options).tap do |data|
                break JSON.dump(data)
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
