module Dapp
  class Builder
    # Manual tagging strategy
    module ManualTagging
      def tag_manual(image_id)
        return unless opts[:tags]

        log 'Applying manual tags'

        opts[:tags].each do |tag|
          spec = {
            name: name,
            tag: tag,
            registry: opts[:docker_registry]
          }

          docker.tag image_id, spec
        end
      end
    end
  end
end
