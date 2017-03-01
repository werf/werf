module Dapp
  # Dimg
  module Dimg
    module Mod
      # Tags
      module Tags
        protected

        def git_repo
          @git_repo ||= GitRepo::Own.new(self)
        end

        def tags
          tags = simple_tags + branch_tags + commit_tags + build_tags + ci_tags
          tags << :latest if tags.empty?
          tags
        end

        def simple_tags
          dapp.cli_options[:tag]
        end

        def branch_tags
          return [] unless dapp.cli_options[:tag_branch]
          raise Error::Dimg, code: :git_branch_without_name if (branch = git_repo.branch) == 'HEAD'
          [branch]
        end

        def commit_tags
          return [] unless dapp.cli_options[:tag_commit]
          commit = git_repo.latest_commit
          [commit]
        end

        def build_tags
          return [] unless dapp.cli_options[:tag_build_id]

          if ENV['GITLAB_CI']
            build_id = ENV['CI_BUILD_ID']
          elsif ENV['TRAVIS']
            build_id = ENV['TRAVIS_BUILD_NUMBER']
          else
            raise Error::Dimg, code: :ci_environment_required
          end

          [build_id]
        end

        def ci_tags
          return [] unless dapp.cli_options[:tag_ci]

          if ENV['GITLAB_CI']
            branch = ENV['CI_BUILD_REF_NAME']
            tag = ENV['CI_BUILD_TAG']
          elsif ENV['TRAVIS']
            branch = ENV['TRAVIS_BRANCH']
            tag = ENV['TRAVIS_TAG']
          else
            raise Error::Dimg, code: :ci_environment_required
          end

          [branch, tag].compact
        end
      end # Tags
    end # Mod
  end # Dimg
end # Dapp
