module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      SUBCOMMANDS = ['build', 'push', 'spush', 'list', 'run', 'stages', 'cleanup', 'bp', 'mrproper', 'stage image', 'tag', 'build-context', 'cleanup repo', 'flush repo'].freeze

      banner <<BANNER.freeze
Usage: dapp dimg [options] subcommand [subcommand options]

Available subcommands: (for details, dapp dimg SUB-COMMAND --help)

  dapp dimg build [options] [DIMG ...]
  dapp dimg bp [options] [DIMG ...] REPO
  dapp dimg push [options] [DIMG ...] REPO
  dapp dimg spush [options] [DIMG] REPO
  dapp dimg tag [options] [DIMG] TAG
  dapp dimg list [options] [DIMG ...]
  dapp dimg run [options] [DIMG] [DOCKER ARGS]
  dapp dimg cleanup repo [options] [DIMG ...] REPO
  dapp dimg flush repo [options] [DIMG ...] REPO
  dapp dimg cleanup [options]
  dapp dimg mrproper [options]
  dapp dimg stage image [options] [DIMG]
  dapp dimg stages
  dapp dimg build-context

Options:
BANNER
    end
  end
end
