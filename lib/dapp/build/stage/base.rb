module Dapp
  module Build
    module Stage
      class Base
        include CommonHelper

        attr_accessor :prev_stage, :next_stage
        attr_reader :build

        def initialize(build, relative_stage)
          @build = build

          @next_stage = relative_stage
          @next_stage.prev_stage = self
        end

        def name
          raise
        end

        def do_build
          return if image_exist?
          prev_stage.do_build if prev_stage
          build_image!
        end

        def image_exist?
          build.docker.image_exist? image_name
        end

        def build_image!
          build.log self.class.to_s
          build.docker.build_image! image: image, name: image_name
        end

        def image
          @image ||= begin
            Image.new(from: from_image_name).tap do |image|
              volumes = ["#{build.build_path}:#{build.container_build_path}"]
              volumes << "#{build.local_git_artifact.repo.dir_path}:#{build.local_git_artifact.repo.container_build_dir_path}" if build.local_git_artifact
              image.build_opts! volume: volumes
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

        def signature
          hashsum prev_stage.signature
        end
      end # Base
    end # Stage
  end # Build
end # Dapp
