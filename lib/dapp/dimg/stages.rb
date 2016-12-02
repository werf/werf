module Dapp
  # Dimg
  class Dimg
    # Stages
    module Stages
      def signature
        last_stage.send(:signature)
      end

      def stage_cache_format
        "#{project.stage_cache}:%{signature}"
      end

      def stage_dapp_label
        project.stage_dapp_label
      end

      def all_images
        @all_images ||= all_stages.map(&:image).uniq!(&:name)
      end

      protected

      def last_stage
        @last_stage ||= if scratch?
                          Build::Stage::ImportArtifact.new(self)
                        else
                          Build::Stage::DockerInstructions.new(self)
                        end
      end

      def export_images
        all_images.select(&:tagged?)
      end

      def import_images
        all_images.select { |image| !image.tagged? }
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

      def artifacts_stages
        @artifacts_stages ||= stages.select { |stage| stage.artifact? }
      end

      def all_stages
        stages + artifacts.map(&:all_stages).flatten
      end
    end # Stages
  end # Dimg
end # Dapp
