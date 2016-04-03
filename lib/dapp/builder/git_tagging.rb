module Dapp
  class Builder
    # Git tagging strategy
    module GitTagging
      def tag_git(image_id)
        spec = {
          name: name,
          registry: opts[:docker_registry]
        }

        { commit: 'rev-parse HEAD', branch: 'rev-parse --abbrev-ref HEAD' }.each do |tag_type, git_command|
          next unless opts[:"tag_#{tag_type}"]

          log "Applying git #{tag_type} tag"

          docker.tag image_id, spec.merge(tag: shellout("git -C #{home_path} #{git_command}").stdout.strip)
        end
      end
    end
  end
end
