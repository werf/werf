module Dapp
  module Dimg
    class Dimg
      module Stages
        def signature
          last_stage.signature
        end

        def stage_by_name(name)
          stages.find { |s| s.name == name }
        end

        def stage_cache_format
          "#{dapp.stage_cache}:%{signature}"
        end

        def stage_dapp_label
          dapp.stage_dapp_label
        end

        def tagged_images
          all_images.select(&:tagged?)
        end
        alias export_images tagged_images

        def all_images
          @all_images ||= all_stages.map(&:image).uniq!(&:name)
        end

        def all_stages
          stages + artifacts.map(&:all_stages).flatten
        end

        def last_stage
          @last_stage || begin
            (self.last_stage = last_stage_class.new(self)).tap do |stage|
              dapp.log_secondary_process("#{name || 'nameless'}: calculating stages signatures") do
                stage.signature
              end
            end
          end
        end

        protected

        attr_writer :last_stage

        def last_stage_class
          if scratch?
            Build::Stage::ImportArtifact
          else
            Build::Stage::DockerInstructions
          end
        end

        def import_images
          all_images.select { |image| !image.tagged? }
        end

        def artifacts_stages
          @artifacts_stages ||= stages.select(&:artifact?)
        end

        def stages
          @stages ||= [].tap do |stages|
            stage = last_stage
            loop do
              stages << stage
              break if (stage = stage.prev_stage).nil?
            end
          end
        end
      end # Stages
    end # Mod
  end # Dimg
end # Dapp
