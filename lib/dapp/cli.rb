require 'mixlib/cli'

module Dapp
  # CLI
  class CLI
    include Mixlib::CLI

    banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dappit [options] [PATTERN ...]

    PATTERN                     Applications to process [default: *].

Options:
BANNER

    option :version,
           long: '--version',
           description: 'Show version',
           on: :tail,
           boolean: true,
           proc: proc { puts "Version: #{Dapp::VERSION}" },
           exit: 0

    option :quiet,
           short: '-q',
           long: '--quiet',
           description: 'Suppress logging',
           on: :tail,
           boolean: true,
           proc: proc { Dapp::Builder.default_opts[:log_quiet] = true }

    option :verbose,
           long: '--verbose',
           description: 'Enable verbose output',
           on: :tail,
           boolean: true,
           proc: proc { Dapp::Builder.default_opts[:log_verbose] = true }

    option :help,
           short: '-h',
           long: '--help',
           description: 'Show this message',
           on: :tail,
           boolean: true,
           show_options: true,
           exit: 0

    option :build_dir,
           long: '--build-dir PATH',
           description: 'Build directory',
           proc: proc { |p| Dapp::Builder.default_opts[:build_dir] = p }

    option :dir,
           long: '--dir PATH',
           description: 'Change to directory',
           on: :head

    option :dappfile_name,
           long: '--dappfile-name NAME',
           description: 'Name of Dappfile',
           proc: proc { |n| Dapp::Builder.default_opts[:dappfile_name] = n },
           on: :head

    option :flush_cache,
           long: '--flush-cache',
           description: 'Flush cache',
           boolean: true,
           proc: proc { Dapp::Builder.default_opts[:flush_cache] = true }

    option :docker_registry,
           long: '--docker-registry REGISTRY',
           description: 'Docker registry',
           proc: proc { |r| Dapp::Builder.default_opts[:docker_registry] = r }

    option :cascade_tagging,
           long: '--cascade_tagging',
           description: 'Use cascade tagging',
           boolean: true,
           proc: proc { Dapp::Builder.default_opts[:cascade_tagging] = true }

    option :git_artifact_branch,
           long: '--git-artifact-branch BRANCH',
           description: 'Default branch to archive artifacts from',
           proc: proc { |b| Dapp::Builder.default_opts[:git_artifact_branch] = b }

    def dappfile_path
      @dappfile_path ||= File.join [config[:dir], 'Dappfile'].compact
    end

    def patterns
      @patterns ||= cli_arguments
    end

    def run(argv = ARGV)
      begin
        parse_options(argv)
      rescue OptionParser::InvalidOption => e
        STDERR.puts "Error: #{e.message}"
        exit 1
      end

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
          STDERR.puts "Error! No such app: '#{pattern}' in #{dappfile_path}"
          exit 1
        end
      end
    end

    def process_directory
      Dapp::Builder.default_opts[:shared_build_dir] = true
      patterns.each do |pattern|
        unless Dapp::Builder.process_directory(config[:dir], pattern).any?
          STDERR.puts "Error! No such app '#{pattern}'"
          exit 1
        end
      end
    end
  end
end
