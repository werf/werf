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

      class << self
        def option(name, args)
          if args.delete :builder_opt
            args[:proc] = if args[:boolean]
                            proc { Dapp::Builder.default_opts[name] = true }
                          else
                            proc { |v| Dapp::Builder.default_opts[name] = v }
                          end
          end

          super(name, args)
        end
      end

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

      option :docker_registry,
             long: '--docker-registry REGISTRY',
             description: 'Docker registry',
             builder_opt: true

      option :flush_cache,
             long: '--flush-cache',
             description: 'Flush cache',
             boolean: true,
             builder_opt: true

      option :cascade_tagging,
             long: '--cascade-tagging',
             description: 'Use cascade tagging',
             boolean: true,
             builder_opt: true

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from',
             builder_opt: true

      def dappfile_path
        @dappfile_path ||= File.join [config[:dir], config[:dappfile_name] || 'Dappfile'].compact
      end

      def patterns
        @patterns ||= cli_arguments
      end

      def run(argv = ARGV)
        CLI.parse_options(self, argv)

        patterns << '*' unless patterns.any?

        if File.exist? dappfile_path
          process_file
        else
          process_directory
        end
      end

      def process_file
        patterns.each do |pattern|
          unless Dapp::Builder.process_file(dappfile_path, app_filter: pattern).builded_apps.any?
            STDERR.puts "Error: No such app: '#{pattern}' in #{dappfile_path}"
            exit 1
          end
        end
      end

      def process_directory
        Dapp::Builder.default_opts[:shared_build_dir] = true
        patterns.each do |pattern|
          unless Dapp::Builder.process_directory(config[:dir], pattern).any?
            STDERR.puts "Error: No such app '#{pattern}'"
            exit 1
          end
        end
      end
    end
  end
end
