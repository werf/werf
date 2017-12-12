module Dapp
  class Dapp
    module OptionTags
      def git_local_repo
        @git_repo ||= ::Dapp::Dimg::GitRepo::Own.new(self)
      end

      def tagging_schemes
        %w(git_tag git_branch git_commit custom ci)
      end

      def tags_by_scheme
        @tags_by_scheme_name ||= begin
          [simple_tags, branch_tags, commit_tags, build_tags, ci_tags].reduce({}) do |some_tags_by_scheme, tags_by_scheme|
            tags_by_scheme.in_depth_merge(some_tags_by_scheme)
          end.tap do |tags_by_scheme|
            [:git_branch, :git_tag, :custom].each do |scheme|
              tags_by_scheme[scheme].map!(&method(:consistent_uniq_slugify)) unless tags_by_scheme[scheme].nil?
            end
            tags_by_scheme[:custom] = [:latest] if tags_by_scheme.values.flatten.empty?
          end
        end
      end

      def option_tags
        tags_by_scheme.values.flatten
      end

      def simple_tags
        { custom: options[:tag] }
      end

      def branch_tags
        return {} unless options[:tag_branch]
        raise Error::Dapp, code: :git_branch_without_name if (branch = git_local_repo.branch) == 'HEAD'
        { git_branch: [branch] }
      end

      def commit_tags
        return {} unless options[:tag_commit]
        { git_commit: [git_local_repo.latest_commit] }
      end

      def build_tags
        return {} unless options[:tag_build_id]

        if ENV['GITLAB_CI']
          build_id = ENV['CI_BUILD_ID']
        elsif ENV['TRAVIS']
          build_id = ENV['TRAVIS_BUILD_NUMBER']
        else
          raise Error::Dapp, code: :ci_environment_required
        end

        { ci: [build_id] }
      end

      def ci_tags
        return {} unless options[:tag_ci]

        {}.tap do |tags_by_scheme|
          if ENV['GITLAB_CI']
            tags_by_scheme[:git_branch] = [ENV['CI_BUILD_REF_NAME']]
            tags_by_scheme[:git_tag]    = [ENV['CI_BUILD_TAG']]
          elsif ENV['TRAVIS']
            tags_by_scheme[:git_branch] = [ENV['TRAVIS_BRANCH']]
            tags_by_scheme[:git_tag]    = [ENV['TRAVIS_TAG']]
          else
            raise Error::Dapp, code: :ci_environment_required
          end

          tags_by_scheme.delete_if { |_, tags| tags.first.nil? }
        end
      end
    end # Tags
  end # Dapp
end # Dapp
