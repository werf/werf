module Dapp
  class CLI
    # CLI bp subcommand
    class Bp < Push
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp bp [options] [APPS PATTERNS] REPO

    APPS PATTERN                Applications to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER
      option :tmp_dir_prefix,
             long: '--tmp-dir-prefix PREFIX',
             description: 'Tmp directory prefix'

      option :lock_timeout,
             long: '--lock-timeout TIMEOUT',
             description: 'Redefine resource locking timeout (in seconds)',
             proc: ->(v) { v.to_i }

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from'

      option :ssh_key,
             long: '--ssh-key SSH_KEY',
             description: ['Enable only specified ssh keys ',
                           '(use system ssh-agent by default)'].join,
             default: nil,
             proc: ->(v) { composite_options(:ssh_key) << v }
    end
  end
end
