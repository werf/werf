module Dapp
  # Project
  class Project
    # Dappfile
    module Dappfile
      def build_configs
        @configs ||= begin
          dimgs(dappfile_path).flatten.tap do |dimgs|
            raise Error::Project, code: :no_such_dimg, data: { dimgs_patterns: dimgs_patterns.join(', ') } if dimgs.empty?
          end
        end
      end

      def dappfile_path
        raise Error::Project, code: :dappfile_not_found unless (dappfile_path = search_up('Dappfile'))
        dappfile_path
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

      def dimgs(dappfile_path)
        Config::DimgGroupMain.new(project: self) do |conf|
          begin
            conf.instance_eval File.read(dappfile_path), dappfile_path
          rescue SyntaxError, StandardError => e
            backtrace = e.backtrace.find { |line| line.start_with?(dappfile_path) }
            message = [NoMethodError, NameError].any? { |err| e.is_a?(err) } ? e.message[/.*(?= for)/] : e.message
            message = "#{backtrace[/.*(?=:in)/]}: #{message}" if backtrace
            raise Error::Dappfile, code: :incorrect, data: { error: e.class.name, message: message }
          end
        end._dimg.select { |dimg| dimgs_patterns.any? { |pattern| dimg._name.nil? || File.fnmatch(pattern, dimg._name) } }.tap do |dimgs|
          dimgs.each { |dimg| dimg.send(:validate!) }
        end
      end
    end # Dappfile
  end # Project
end # Dapp
