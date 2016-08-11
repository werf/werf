require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Build < Base
      include Dapp::Helper::Shellout

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp build [options] [PATTERN ...]

    PATTERN                     Applications to process [default: *].

Options:
BANNER
      option :tmp_dir_prefix,
             long: '--tmp-dir-prefix PREFIX',
             description: 'Tmp directory prefix'

      option :metadata_dir,
             long: '--metadata-dir PATH',
             description: 'Metadata directory'

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Directory where build cache stored (DIR/.dapps-build by default)'

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
