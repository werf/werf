module Dapp
  module Dimg
    module DockerRegistry
      class Default < Base
        DEFAULT_HOSTNAME_URL = 'https://registry.hub.docker.com'.freeze

        def initialize(repo, repo_suffix)
          super(repo, DEFAULT_HOSTNAME_URL, repo_suffix)
        end

        def repo_suffix=(val)
          val = "library/#{val}" if val.split('/').one?
          super(val)
        end
      end
    end
  end # Dimg
end # Dapp
