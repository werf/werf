module Dapp
  module Config
    module Directive
      # GitArtifactLocal
      class GitArtifactLocal < ArtifactBase
        protected

        alias add export
        undef_method :export
      end
    end
  end
end
