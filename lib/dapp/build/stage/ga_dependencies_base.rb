module Dapp
  module Build
    module Stage
      # GADependenciesBase
      class GADependenciesBase < Base
        def prepare_image
          super
          dimg.git_artifacts.each do |git_artifact|
            image.add_service_change_label(git_artifact.full_name.to_sym => git_artifact.latest_commit)
          end
        end

        def empty?
          dimg.git_artifacts.empty? ? true : false
        end
      end # GADependenciesBase
    end # Stage
  end # Build
end # Dapp
