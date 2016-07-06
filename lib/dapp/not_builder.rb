module Dapp
  class NotBuilder
    attr_reader :cli_options, :patterns

    def initialize(cli_options:, patterns: nil)
      @cli_options = cli_options

      @patterns = patterns || []
      @patterns << '*' unless @patterns.any?

      build_confs
    end

    def build
      @build_confs.each do |build_conf|
        puts build_conf.name
        options = { conf: build_conf, opts: cli_options }
        Application.new(**options).build_and_fixate!
      end
    end

    private

    def build_confs
      options = {}
      [:log_quiet, :log_verbose, :type].each { |opt| options[opt] = cli_options[opt] }

      @build_confs = if File.exist? dappfile_path
        process_file(**options)
      else
        process_directory(**options)
      end
    end

    def process_file(**options)
      patterns.map do |pattern|
        unless (apps = Loader.process_file(dappfile_path, app_filter: pattern, **options)).any?
          STDERR.puts "Error: No such app: '#{pattern}' in #{dappfile_path}"
          exit 1
        end
        apps
      end.flatten
    end

    def process_directory(**options)
      options[:shared_build_dir] = true
      patterns.map do |pattern|
        unless (apps = Loader.process_directory(cli_options[:dir], pattern, **options)).any?
          STDERR.puts "Error: No such app '#{pattern}'"
          exit 1
        end
        apps
      end.flatten
    end

    def dappfile_path
      @dappfile_path ||= File.join [cli_options[:dir], cli_options[:dappfile_name] || 'Dappfile'].compact
    end
  end # NotBuilder
end # Dapp
