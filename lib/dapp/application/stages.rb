module Dapp
  # Application
  class Application
    # Stages
    module Stages
      def signature
        last_stage.send(:signature)
      end

      def stage_cache_format
        "#{project.cache_format % { application_name: config._basename }}:%{signature}"
      end

      def stage_dapp_label
        project.stage_dapp_label_format % { application_name: config._basename }
      end

      def images
        (@images ||= []).tap do |images|
          stages.map do |stage|
            if stage.respond_to?(:images)
              images.concat(stage.images)
            else
              images << stage.image
            end
          end.uniq(&:name)
        end
      end

      protected

      attr_reader :last_stage

      def export_images
        images.select(&:tagged?)
      end

      def import_images
        images.select { |image| !image.tagged? }
      end

      def stages
        (@stages ||= []).tap do |stages|
          stage = last_stage
          loop do
            stages << stage
            break if (stage = stage.prev_stage).nil?
          end
        end
      end
    end # Stages
  end # Application
end # Dapp
