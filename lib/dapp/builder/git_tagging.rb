module Dapp
  class Builder
    # Git tagging strategy
    module GitTagging
      def tag_git(image_id)
        spec = {
          name: name,
          registry: opts[:docker_registry]
        }

        if opts[:tag_commit]
          log 'Applying git commit tag'

          docker.tag image_id, spec.merge(tag: shellout("git -C #{home_path} rev-parse HEAD").stdout.strip)
        end

        if opts[:tag_branch]
          log 'Applying git branch tag'

          docker.tag image_id, spec.merge(tag: shellout("git -C #{home_path} rev-parse --abbrev-ref HEAD").stdout.strip)
        end
      end
    end
  end
end
