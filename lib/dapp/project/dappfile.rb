module Dapp
  # Project
  class Project
    # Dappfile
    module Dappfile
      def build_configs
        @configs ||= begin
          dappfiles.map { |dappfile| apps(dappfile, app_filters: apps_patterns) }.flatten.tap do |apps|
            raise Error::Project, code: :no_such_app, data: { apps_patterns: apps_patterns.join(', ') } if apps.empty?
          end
        end
      end

      def dappfiles
        if File.exist?(dappfile_path)                 then [dappfile_path]
        elsif !dapps_dappfiles_pathes.empty?          then dapps_dappfiles_pathes
        elsif (dappfile_path = search_up('Dappfile')) then [dappfile_path]
        else raise Error::Project, code: :dappfile_not_found
        end
      end

      def dappfile_path
        File.join [cli_options[:dir], 'Dappfile'].compact
      end

      def dapps_dappfiles_pathes
        path = []
        path << cli_options[:dir]
        path << '.dapps' unless File.basename(work_dir) == '.dapps'
        path << '*'
        path << 'Dappfile'
        Dir.glob(File.join(path.compact))
      end

      def search_up(file)
        cdir = Pathname(work_dir)
        loop do
          if (path = cdir.join(file)).exist?
            return path.to_s
          end
          break if (cdir = cdir.parent).root?
        end
      end

      def work_dir
        File.expand_path(cli_options[:dir] || Dir.pwd)
      end

      def expand_path(path, number = 1)
        path = File.expand_path(path)
        number.times.each { path = File.dirname(path) }
        path
      end

      def apps(dappfile_path, app_filters:)
        config = Config::Main.new(dappfile_path: dappfile_path, project: self) do |conf|
          begin
            conf.instance_eval File.read(dappfile_path), dappfile_path
          rescue SyntaxError, StandardError => e
            backtrace = e.backtrace.find { |line| line.start_with?(dappfile_path) }
            message = e.is_a?(NoMethodError) ? e.message[/.*(?= for)/] : e.message
            message = "#{backtrace[/.*(?=:in)/]}: #{message}" if backtrace
            raise Error::Dappfile, code: :incorrect, data: { error: e.class.name, message: message }
          end
        end
        config._apps.select { |app| app_filters.any? { |pattern| File.fnmatch(pattern, app._name) } }.tap do |apps|
          apps.each { |app| app.send(:validate!) }
        end
      end
    end # Dappfile
  end # Project
end # Dapp
