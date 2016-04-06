module Dapp
  class Builder
    # CI tagging strategy
    module CiTagging
      def tag_ci(image_id)
        return unless opts[:tag_ci] || opts[:tag_build_id]

        raise 'CI environment required (Travis or GitLab CI)' unless ENV['GITLAB_CI'] || ENV['TRAVIS']

        spec = {
          name: name,
          repo: opts[:docker_repo]
        }

        tag_ci_branch_and_tag image_id, spec
        tag_ci_build_id image_id, spec
      end

      private

      def tag_ci_branch_and_tag(image_id, spec)
        return unless opts[:tag_ci]

        log 'Applying CI tagging strategy'

        if ENV['GITLAB_CI']
          branch = ENV['CI_BUILD_REF_NAME']
          tag = ENV['CI_BUILD_TAG']
        elsif ENV['TRAVIS']
          branch = ENV['TRAVIS_BRANCH']
          tag = ENV['TRAVIS_TAG']
        end

        docker.tag image_id, spec.merge(tag: tag) if tag
        docker.tag image_id, spec.merge(tag: branch) if branch
      end

      def tag_ci_build_id(image_id, spec)
        return unless opts[:tag_build_id]

        log 'Applying CI build id tag'

        if ENV['GITLAB_CI']
          build_id = ENV['CI_BUILD_ID']
        elsif ENV['TRAVIS']
          build_id = ENV['TRAVIS_BUILD_NUMBER']
        end

        docker.tag image_id, spec.merge(tag: build_id) if build_id
      end
    end
  end
end
