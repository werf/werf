module Dapp
  class Dapp
    module OptionTags
      def git_local_repo
        @git_repo ||= ::Dapp::Dimg::GitRepo::Own.new(self)
      end

      def option_tags
        @tags ||= begin
          tags = simple_tags + branch_tags + commit_tags + build_tags + ci_tags
          tags << :latest if tags.empty?
          tags
        end
      end

      def simple_tags
        options[:tag]
      end

      def branch_tags
        return [] unless options[:tag_branch]
        raise Error::Dapp, code: :git_branch_without_name if (branch = git_local_repo.branch) == 'HEAD'
        [branch]
      end

      def commit_tags
        return [] unless options[:tag_commit]
        commit = git_local_repo.latest_commit
        [commit]
      end

      def build_tags
        return [] unless options[:tag_build_id]

        if ENV['GITLAB_CI']
          build_id = ENV['CI_BUILD_ID']
        elsif ENV['TRAVIS']
          build_id = ENV['TRAVIS_BUILD_NUMBER']
        else
          raise Error::Dapp, code: :ci_environment_required
        end

        [build_id]
      end

      def ci_tags
        return [] unless options[:tag_ci]

        if ENV['GITLAB_CI']
          branch = ENV['CI_BUILD_REF_NAME']
          tag = ENV['CI_BUILD_TAG']
        elsif ENV['TRAVIS']
          branch = ENV['TRAVIS_BRANCH']
          tag = ENV['TRAVIS_TAG']
        else
          raise Error::Dapp, code: :ci_environment_required
        end

        [branch, tag].compact
      end
    end # Tags
  end # Dapp
end # Dapp
