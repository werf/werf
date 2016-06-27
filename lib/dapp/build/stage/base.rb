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

        def build!
          return if image.exist?
          prev_stage.build! if prev_stage
          build.log self.class.to_s
          image.build!
        end

        def fixate!
          prev_stage.fixate! if prev_stage
          image.fixate!
        end

        def signature
          hashsum prev_stage.signature
        end

        def image
          @image ||= begin
            DockerImage.new(name: image_name, from: from_image).tap do |image|
              # FIXME only in source stages
              image.add_volume "#{build.build_path}:#{build.container_build_path}"
              image.add_volume "#{build.local_git_artifact.repo.dir_path}:#{build.local_git_artifact.repo.container_build_dir_path}" if build.local_git_artifact
              yield image if block_given?
            end
          end
        end

        protected

        def name
          self.class.to_s.split('::').last.split(/(?=[[:upper:]]|[0-9])/).join(?_).downcase.to_sym
        end

        def from_image
          prev_stage.image if prev_stage || begin
            raise 'missing from_image'
          end
        end

        def image_name
          "dapp:#{signature}"
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
