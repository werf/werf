module Dapp
  module Build
    module Stage
      class Source2 < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = AppInstall.new(build, self)
          super
        end

        def name
          :source_2
        end

        def signature
          hashsum [prev_stage.signature,
                   *build.infra_setup_commands,
                   *commit_list] # TODO chef
        end

        def git_artifact_signature
          hashsum [prev_stage.signature,
                   *build.infra_setup_commands]
        end
      end # Source2
    end # Stage
  end # Build
end # Dapp
