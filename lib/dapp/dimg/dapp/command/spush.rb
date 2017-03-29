module Dapp
  module Dimg
    module Dapp
      module Command
        module Spush
          def spush(repo)
            validate_repo_name(repo)
            one_dimg!
            Dimg.new(config: build_configs.first, dapp: self, ignore_git_fetch: true, should_be_built: true).tap do |dimg|
              dimg.export!(repo, format: spush_format)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
