module Dapp
  module Build
    module Stage
      # From
      class From < Base
        def signature
          hashsum [from_image_name, application.config._docker._from_cache_version]
        end

        def build!
          return if image.tagged? && !application.show_only
          from_image.pull! unless application.show_only
          build_log
          image.build!(application.logging?) unless application.show_only
        end

        def save_in_cache!
          super
          from_image.untag! if from_image.pulled? && from_image.tagged? && !application.show_only
        end

        private

        def from_image_name
          application.config._docker._from
        end

        def from_image
          StageImage.new(name: from_image_name)
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
