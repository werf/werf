module Dapp
  module Build
    module Stage
      # From
      class From < Base
        def signature
          hashsum [from_image_name, application.config._docker._from_cache_version]
        end

        def build!
          return if image.tagged? && !application.dry_run
          log_build_time do
            from_image.pull! unless application.dry_run
            log_build
            image.build!(application.logging?) unless application.dry_run
          end
        end

        def save_in_cache!
          super
          from_image.untag! if from_image.pulled? && from_image.tagged? && !application.dry_run
        end

        private

        def from_image_name
          application.config._docker._from
        end

        def from_image
          StageImage.new(name: from_image_name)
        end

        def image_info
          image.info
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
