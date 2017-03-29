module Dapp
  class Dapp
    module Dappfile
      def local_git_artifact_exclude_paths(&blk)
        super do |exclude_paths|
          exclude_paths << 'Dappfile'

          yield exclude_paths if block_given?
        end
      end

      def dappfile_path
        raise Error::Dapp, code: :dappfile_not_found unless (dappfile_path = search_file_upward('Dappfile'))
        dappfile_path
      end

      def work_dir
        File.expand_path(cli_options[:dir] || Dir.pwd)
      end

      def expand_path(path, number = 1)
        path = File.expand_path(path)
        number.times.each { path = File.dirname(path) }
        path
      end

      def config
        @config ||= begin
          ::Dapp::Config::Config.new(dapp: self).tap do |config|
            begin
              config.instance_eval File.read(dappfile_path), dappfile_path
              config.validate!
            rescue SyntaxError, StandardError => e
              backtrace = e.backtrace.find { |line| line.start_with?(dappfile_path) }
              message = begin
                case e
                when NoMethodError
                  e.message =~ /`.*'/
                  "undefined method #{Regexp.last_match}"
                when NameError then e.message[/.*(?= for)/]
                else
                  e.message
                end
              end
              message = "#{backtrace[/.*(?=:in)/]}: #{message}" if backtrace
              raise Error::Dappfile, code: :incorrect, data: { error: e.class.name, message: message }
            end
          end
        end
      end
    end # Dappfile
  end # Dapp
end # Dapp
