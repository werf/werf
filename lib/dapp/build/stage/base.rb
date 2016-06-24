module Dapp
  module Build
    module Stage
      class Base
        include CommonHelper

        attr_accessor :prev_stage, :next_stage

        # FIXME rename Build class to smth else
        attr_reader :build

        def initialize(build, next_stage)
          @build = build

          @next_stage = next_stage
          @next_stage.prev_stage = self
        end

        def name
          self.class.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join(?_).downcase.to_sym
        end

        def build!
          return if image_exist?
          prev_stage.build! if prev_stage
          build.log self.class.to_s
          puts image.bash_commands.join
          build.docker.build_image! image_specification: image, image_name: image_name
        end

        def signature
          hashsum prev_stage.signature
        end

        protected

        def image
          @image ||= begin
            ImageSpecification.new(from_name: from_image_name).tap do |image|
              image.add_volume "#{build.build_path}:#{build.container_build_path}"
              image.add_volume "#{build.local_git_artifact.repo.dir_path}:#{build.local_git_artifact.repo.container_build_dir_path}" if build.local_git_artifact
              yield image if block_given?
            end
          end
        end

        def from_image_name
          @from_image_name || (prev_stage.image_name if prev_stage) || begin
            raise 'missing from_image_name'
          end
        end

        def image_name
          "dapp:#{signature}"
        end

        def image_exist?
          build.docker.image_exist? image_name
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
