module Dapp
  class Builder
    # Cascade tagging strategy
    module CascadeTagging
      # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      def tag_cascade(image_id)
        return unless opts[:tag_cascade]

        log 'Applying cascade tagging'

        opts[:build_history_length] ||= 10

        spec = {
          name: name,
          tag: home_branch,
          repo: opts[:docker_repo]
        }

        # return if nothing changed
        return if image_id == docker.image_id(spec)

        # remove excess tags
        tags_to_remove = docker.images(name: spec[:name], repo: spec[:repo])
        tags_to_remove.map! { |image| image[:tag] }
        tags_to_remove.select! { |tag| tag.start_with?("#{spec[:tag]}_") && tag.sub(/^#{spec[:tag]}_/, '').to_i >= opts[:build_history_length] }
        tags_to_remove.each do |tag_to_remove|
          docker.rmi spec.merge(tag: tag_to_remove)
        end

        # shift old images: 1 -> 2, 2 -> 3, ..., n -> n+1
        (opts[:build_history_length] - 1).downto(1).each do |n|
          origin = spec.merge(tag: "#{spec[:tag]}_#{n}")

          if docker.image_exist?(**origin)
            docker.tag origin, spec.merge(tag: "#{spec[:tag]}_#{n + 1}")
          end
        end

        # shift top -> 1
        docker.tag spec, spec.merge(tag: "#{spec[:tag]}_1") if docker.image_exist?(**spec)

        # tag top
        docker.tag image_id, spec
      end
      # rubocop:enable Metrics/AbcSize, Metrics/MethodLength
    end
  end
end
