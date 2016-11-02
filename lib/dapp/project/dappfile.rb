module Dapp
  # Project
  class Project
    # Dappfile
    module Dappfile
      def build_configs
        @configs ||= begin
          dappfiles.map { |dappfile| dimgs(dappfile, dimgs_filters: dimgs_patterns) }.flatten.tap do |dimgs|
            raise Error::Project, code: :no_such_dimg, data: { dimgs_patterns: dimgs_patterns.join(', ') } if dimgs.empty?
          end
        end
      end

      def dappfiles
        if File.exist?(dappfile_path)                 then [dappfile_path]
        elsif !dimgs_dappfiles_pathes.empty?          then dimgs_dappfiles_pathes
        elsif (dappfile_path = search_up('Dappfile')) then [dappfile_path]
        else raise Error::Project, code: :dappfile_not_found
        end
      end

      def dappfile_path
        File.join [cli_options[:dir], 'Dappfile'].compact
      end

      def dimgs_dappfiles_pathes
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

      def dimgs(dappfile_path, dimgs_filters:)
        Config::DimgGroupMain.new(project: self) do |conf|
          begin
            conf.instance_eval File.read(dappfile_path), dappfile_path
          rescue SyntaxError, StandardError => e
            backtrace = e.backtrace.find { |line| line.start_with?(dappfile_path) }
            message = e.is_a?(NoMethodError) ? e.message[/.*(?= for)/] : e.message
            message = "#{backtrace[/.*(?=:in)/]}: #{message}" if backtrace
            raise Error::Dappfile, code: :incorrect, data: { error: e.class.name, message: message }
          end
        end
        ._dimg.select { |dimg| dimgs_filters.any? { |pattern| File.fnmatch(pattern, dimg._name.to_s) } }.tap do |dimgs|
          dimgs.each { |dimg| dimg.send(:validate!) }
        end
      end
    end # Dappfile
  end # Project
end # Dapp
