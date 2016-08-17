require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Build < Base
      include Dapp::Helper::Shellout

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp build [options] [APPS PATTERN ...]

    APPS PATTERN                Applications to process [default: *].

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

      option :introspect_error,
             long: '--introspect-error',
             boolean: true,
             default: false

      option :introspect_before_error,
             long: '--introspect-before-error',
             boolean: true,
             default: false

      option :introspect_stage,
             long: '--introspect-stage STAGE',
             proc: proc { |v| v.to_sym },
             in: [nil, :from, :infra_install, :source_1_archive, :source_1, :install, :artifact, :source_2,
                  :infra_setup, :source_3, :chef_cookbooks, :setup, :source_4, :source_5, :docker_instructions]
    end
  end
end
