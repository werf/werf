module Dapp
  module Build
    module Stage
      # From
      class From < Base
        def signature
          hashsum [*dependencies.flatten]
        end

        def dependencies
          [from_image_name, application.config._docker._from_cache_version, Dapp::BUILD_CACHE_VERSION]
        end

        protected

        def image_build
          from_image.pull!(log_verbose: application.project.log_verbose?, log_time: application.project.log_time?)
          raise Error::Build, code: :from_image_not_found, data: { name: from_image_name } unless from_image.tagged?
          super
        end

        def should_not_be_detailed?
          from_image.tagged?
        end

        private

        def from_image_name
          application.config._docker._from
        end

        def from_image
          @from_image ||= begin
            if from_image_name.nil?
              Image::Scratch.new(project: application.project)
            else
              Image::Stage.new(name: from_image_name, project: application.project)
            end
          end
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
