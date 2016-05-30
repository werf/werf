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
             builder_opt: true,
             on: :head

      option :build_dir,
             long: '--build-dir PATH',
             description: 'Build directory',
             builder_opt: true

      option :docker_repo,
             long: '--docker-repo REPO',
             description: 'Docker repo',
             builder_opt: true

      option :docker_socket,
             long: '--docker-socket SOCKET',
             description: 'Docker socket'

      option :flush_cache,
             long: '--flush-cache',
             description: 'Flush cache',
             boolean: true,
             builder_opt: true

      option :tag_cascade,
             long: '--tag-cascade',
             description: 'Use cascade tagging',
             boolean: true,
             builder_opt: true

      option :tag_ci,
             long: '--tag-ci',
             description: 'Tag by CI branch and tag',
             boolean: true,
             builder_opt: true

      option :tag_build_id,
             long: '--tag-build-id',
             description: 'Tag by CI build id',
             boolean: true,
             builder_opt: true

      option :tag,
             long: '--tag TAG',
             description: 'Add tag (can be used one or more times)',
             proc: proc { |v| composite_options(:tags) << v }

      option :tag_commit,
             long: '--tag-commit',
             description: 'Tag by git commit',
             boolean: true,
             builder_opt: true

      option :tag_branch,
             long: '--tag-branch',
             description: 'Tag by git branch',
             boolean: true,
             builder_opt: true

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from',
             builder_opt: true

      def self.composite_options(opt)
        @composite_options ||= {}
        @composite_options[opt] ||= []
      end

      def dappfile_path
        @dappfile_path ||= File.join [config[:dir], config[:dappfile_name] || 'Dappfile'].compact
      end

      def patterns
        @patterns ||= cli_arguments
      end

      def run(argv = ARGV)
        CLI.parse_options(self, argv)

        patterns << '*' unless patterns.any?

        # TODO: creating build_conf
        build_conf = {
          from: 'ubuntu:14.04',
          git_artifact: {
            local: {where_to_add: '/myapp'},
          },
        }

        # builder options
        options = {
            docker: Dapp::Docker.new(socket: config[:docker_socket]),
            conf: build_conf,
            opts: config
        }

        # TODO: choosing builder
        builder = if 1 == 2
                    Dapp::Builder::Chef
                  else
                    Dapp::Builder::Shell
                  end
        builder.new(**options).run
      end
    end
  end
end
