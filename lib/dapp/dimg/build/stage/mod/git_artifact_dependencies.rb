module Dapp
  module Dimg
    module Build
      module Stage
        module Mod
          module GitArtifactsDependencies
            def local_git_artifacts_dependencies
              dimg.local_git_artifacts.map do |git_artifact|
                args = []
                args << self
                if dimg.dev_mode?
                  args << git_artifact.latest_commit
                  args << nil
                end
                git_artifact.stage_dependencies_checksums(*args)
              end
            end
          end
        end # Mod
      end # Stage
    end # Build
  end # Dimg
end # Dapp
