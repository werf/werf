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

        def save_in_cache!
          from_image.untag! if from_image.pulled?
          super
        end

        protected

        def image_do_build
          from_image.pull!(log_time: application.log_time?)
          raise Error::Build, code: :from_image_not_found, data: { name: from_image_name } if from_image.built_id.nil?
          super
        end

        private

        def from_image_name
          application.config._docker._from
        end

        def from_image
          @from_image ||= Image::Stage.new(name: from_image_name)
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
