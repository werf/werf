module Dapp
  class Dapp
    module Dappfile
      module Error
        class DappfileYmlErrorResponse < ::Dapp::Error::Default
          def initialize(error_code, response)
            net_status = {}
            net_status[:code] = error_code
            net_status[:message] = response["message"] if response["message"]
            super(net_status)
          end
        end
      end

      def local_git_artifact_exclude_paths(&blk)
        super do |exclude_paths|
          exclude_paths << 'Dappfile'
          exclude_paths << "dappfile.yml"
          exclude_paths << "dappfile.yaml"

          yield exclude_paths if block_given?
        end
      end

      def expand_path(path, number = 1)
        path = File.expand_path(path)
        number.times.each { path = File.dirname(path) }
        path
      end

      def dappfile_exists?
        File.exist?(path("dappfile.yml")) ||
          File.exist?(path("dappfile.yaml")) ||
            File.exist?(path("Dappfile")) ||
              ENV["DAPP_LOAD_CONFIG_PATH"]
      end

      def config
        @config ||= begin
          config = nil

          dappfile_yml = path("dappfile.yml").to_s
          dappfile_yaml = path("dappfile.yaml").to_s
          dappfile_ruby = path("Dappfile").to_s

          if ENV["DAPP_LOAD_CONFIG_PATH"]
            config = YAML.load_file ENV["DAPP_LOAD_CONFIG_PATH"]
          elsif File.exist? dappfile_yml
            config = load_dappfile_yml(dappfile_yml)
          elsif File.exist? dappfile_yaml
            config = load_dappfile_yml(dappfile_yaml)
          elsif File.exist? dappfile_ruby
            config = load_dappfile_ruby(dappfile_ruby)
          else
            raise ::Dapp::Error::Dapp, code: :dappfile_not_found
          end

          if ENV["DAPP_DUMP_CONFIG"]
            puts "-- DAPP_DUMP_CONFIG BEGIN"
            puts YAML.dump(config)
            puts "-- DAPP_DUMP_CONFIG END"
          end

          config
        end # begin
      end

      def load_dappfile_ruby(dappfile_path)
        ::Dapp::Config::Config.new(dapp: self).tap do |config|
          begin
            config.instance_eval File.read(dappfile_path), dappfile_path
            config.after_parsing!
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
            raise ::Dapp::Error::Dappfile, code: :incorrect, data: { error: e.class.name, message: message }
          end # begin-rescue
        end
      end

      def load_dappfile_yml(dappfile_path)
        if dappfile_yml_bin_path = ENV["DAPP_BIN_DAPPFILE_YML"]
          unless File.exists? dappfile_yml_bin_path
            raise ::Dapp::Error::Dapp, code: :dappfile_yml_bin_path_not_found, data: {path: dappfile_yml_bin_path}
          end
        else
          dappfile_yml_bin_path = File.join(::Dapp::Dapp.home_dir, "bin", "dappfile-yml", ::Dapp::VERSION, "dappfile-yml")
          unless File.exists? dappfile_yml_bin_path
            download_dappfile_yml_bin(dappfile_yml_bin_path)
          end
        end

        cmd_res = shellout "#{dappfile_yml_bin_path} -dappfile #{dappfile_path}"

        raw_json_response = nil
        if cmd_res.exitstatus == 0
          raw_json_response = cmd_res.stdout
        elsif cmd_res.exitstatus == 16
          raw_json_response = cmd_res.stderr
        else
          shellout_cmd_should_succeed! cmd_res
        end

        response = JSON.parse(raw_json_response)

        raise ::Dapp::Dapp::Error::DappfileYmlErrorResponse.new(response["error"], response) if response["error"]

        YAML.load response["dappConfig"]
      end

      def download_dappfile_yml_bin(dappfile_yml_bin_path)
        lock("downloader.bin.dappfile-yml", default_timeout: 1800) do
          return if File.exists? dappfile_yml_bin_path

          log_process("Downloading dappfile-yml dapp dependency") do
            location = URI("https://dl.bintray.com/dapp/ruby2go/#{::Dapp::VERSION}/dappfile-yml")

            tmp_bin_path = File.join(self.class.tmp_base_dir, "dappfile-yml-#{SecureRandom.uuid}")
            ::Dapp::Downloader.download(location, tmp_bin_path, show_progress: true, progress_titile: dappfile_yml_bin_path)

            checksum_location = URI("https://dl.bintray.com/dapp/ruby2go/#{::Dapp::VERSION}/dappfile-yml.sha")
            tmp_bin_checksum_path = tmp_bin_path + ".checksum"
            ::Dapp::Downloader.download(checksum_location, tmp_bin_checksum_path)

            if Digest::SHA256.hexdigest(File.read(tmp_bin_path)) != File.read(tmp_bin_checksum_path).strip
              raise ::Dapp::Error::Dapp, code: :download_failed_bad_dappfile_yml_checksum, data: {url: location.to_s, checksum_url: checksum_location.to_s}
            end

            File.chmod(0755, tmp_bin_path)
            FileUtils.mkdir_p File.dirname(dappfile_yml_bin_path)
            FileUtils.mv tmp_bin_path, dappfile_yml_bin_path
          end # log_process
        end # lock
      end

    end # Dappfile
  end # Dapp
end # Dapp
