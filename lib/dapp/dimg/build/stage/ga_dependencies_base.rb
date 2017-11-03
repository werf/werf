module Dapp
  module Dimg
    module Build
      module Stage
        class GADependenciesBase < Base
          def prepare_image
            super do
              dimg.git_artifacts.each do |git_artifact|
                image.add_service_change_label("dapp-git-#{git_artifact.paramshash}-commit".to_sym => git_artifact.latest_commit)
              end
            end
          end

          def empty?
            dimg.git_artifacts.empty? || super
          end

          def image_should_be_untagged_condition
            return false unless image.tagged?
            dimg.git_artifacts.any? do |git_artifact|
              !git_artifact.repo.commit_exists? image.labels["dapp-git-#{git_artifact.paramshash}-commit"]
            end
          end
        end # GADependenciesBase
      end # Stage
    end # Build
  end # Dimg
end # Dapp
