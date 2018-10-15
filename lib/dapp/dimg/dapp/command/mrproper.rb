module Dapp
  module Dimg
    module Dapp
      module Command
        module Mrproper
          def mrproper
            ruby2go_cleanup_command(:reset, ruby2go_cleanup_reset_options_dump)
          end

          def ruby2go_cleanup_reset_options_dump
            cleanup_repo_proper_cache_version_options.merge(
              mode: {
                all: !!options[:proper_all],
                dev_mode_cache: !!options[:proper_dev_mode_cache],
                cache_version: proper_cache_version?
              },
              common_options: {
                dry_run: dry_run?
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
