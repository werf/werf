require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Build < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp build [options] [PATTERN ...]

    PATTERN                     Applications to process [default: *].

Options:
BANNER

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Build directory'

      option :build_cache_dir,
             long: '--build-cache-dir PATH',
             description: 'Build cache directory'

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from'

      option :show_only,
             long: '--show-only',
             boolean: true
    end
  end
end
