require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Build
      include Mixlib::CLI

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp build [options] [PATTERN ...]

    PATTERN                     Applications to process [default: *].

Options:
BANNER

      option :log_quiet,
             short: '-q',
             long: '--quiet',
             description: 'Suppress logging',
             on: :tail,
             boolean: true,
             builder_opt: true

      option :log_verbose,
             long: '--verbose',
             description: 'Enable verbose output',
             on: :tail,
             boolean: true,
             builder_opt: true

      option :help,
             short: '-h',
             long: '--help',
             description: 'Show this message',
             on: :tail,
             boolean: true,
             show_options: true,
             exit: 0

      option :dir,
             long: '--dir PATH',
             description: 'Change to directory',
             on: :head

      option :dappfile_name,
             long: '--dappfile-name NAME',
             description: 'Name of Dappfile',
             on: :head

      option :type,
             long: '--type NAME',
             description: 'type of Dappfile',
             default: :chef,
             proc: proc { |opt| opt.to_sym },
             on: :head

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Build directory'

      option :docker_repo,
             long: '--docker-repo REPO',
             description: 'Docker repo'

      option :docker_socket,
             long: '--docker-socket SOCKET',
             description: 'Docker socket'

      option :flush_cache,
             long: '--flush-cache',
             description: 'Flush cache'

      option :tag_cascade,
             long: '--tag-cascade',
             description: 'Use cascade tagging',
             boolean: true

      option :tag_ci,
             long: '--tag-ci',
             description: 'Tag by CI branch and tag',
             boolean: true

      option :tag_build_id,
             long: '--tag-build-id',
             description: 'Tag by CI build id',
             boolean: true

      option :tag,
             long: '--tag TAG',
             description: 'Add tag (can be used one or more times)',
             proc: proc { |v| composite_options(:tags) << v }

      option :tag_commit,
             long: '--tag-commit',
             description: 'Tag by git commit',
             boolean: true

      option :tag_branch,
             long: '--tag-branch',
             description: 'Tag by git branch',
             boolean: true

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from'

      def self.composite_options(opt)
        @composite_options ||= {}
        @composite_options[opt] ||= []
      end

      def run(argv = ARGV)
        CLI.parse_options(self, argv)
        NotBuilder.new(cli_options: config, patterns: cli_arguments).build
      end
    end
  end
end
