module Buildit
  class Builder
    module CascadeTagging
      def tag_cascade(image_id)
        return unless opts[:cascade_tagging]

        log "Applying cascade tagging"

        opts[:build_history_length] ||= 10

        i = {
          name: name,
          tag: home_branch,
          registry: opts[:docker_registry]
        }

        # return if nothing changed
        return if image_id == docker.image_id(i)

        # remove excess tags
        tags_to_remove = docker.images(name: i[:name], registry: i[:registry])
          .map{|image| image[:tag] }
          .select{|tag| tag.start_with?("#{i[:tag]}_") && tag.sub(/^#{i[:tag]}_/, '').to_i >= opts[:build_history_length] }
        tags_to_remove.each do |tag_to_remove|
          puts "TROMPUMPUM!"
          docker.rmi i.merge(tag: tag_to_remove)
        end

        # shift old images: 1 -> 2, 2 -> 3, ..., n -> n+1
        (opts[:build_history_length] - 1).downto(1).each do |n|
          origin = i.merge(tag: "#{i[:tag]}_#{n}")

          if docker.image_exists? **origin
            docker.tag origin, i.merge(tag: "#{i[:tag]}_#{n + 1}")
          end
        end

        # shift top -> 1
        if docker.image_exists? **i
          docker.tag i, i.merge(tag: "#{i[:tag]}_1")
        end

        # tag top
        docker.tag image_id, i
      end
    end
  end
end
