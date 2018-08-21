module Dapp
  class Dapp
    module Ruby2Go
      def ruby2go_image(args_hash)
        _ruby2go("image", args_hash)
      end

      def ruby2go_builder(args_hash)
        _ruby2go("builder", args_hash)
      end

      def ruby2go_git_artifact(args_hash)
        _ruby2go("git-artifact", args_hash)
      end

      def ruby2go_dappdeps(args_hash)
        _ruby2go("dappdeps", args_hash)
      end

      def ruby2go_git_repo(args_hash)
        _ruby2go("git-repo", args_hash)
      end

      def ruby2go_init
        @_call_after_before_terminate << proc {
          FileUtils.rmtree(@_ruby2go_tmp_dir) if @_ruby2go_tmp_dir
        }
      end

      def _ruby2go_bin_path_env_var_name(progname)
        "DAPP_BIN_#{progname.gsub("-", "_").upcase}"
      end

      def _ruby2go(progname, args_hash)
        call_id = SecureRandom.uuid

        args_file = File.join(_ruby2go_tmp_dir, "args.#{call_id}.json")
        File.open(args_file, "w") {|f| f.write JSON.dump(args_hash)}

        res_file = File.join(_ruby2go_tmp_dir, "res.#{call_id}.json")

        if bin_path = ENV[_ruby2go_bin_path_env_var_name(progname)]
          unless File.exists? bin_path
            raise ::Dapp::Error::Dapp,
              code: :ruby2go_bin_path_not_found,
              data: {env_var_name: _ruby2go_bin_path_env_var_name(progname), path: bin_path}
          end
        else
          bin_path = File.join(::Dapp::Dapp.home_dir, "bin", progname, ::Dapp::VERSION, progname)
          unless File.exists? bin_path
            _download_ruby2go_bin(progname, bin_path)
          end
        end

        system("#{bin_path} -args-from-file #{args_file} -result-to-file #{res_file}")
        status_code = $?.exitstatus
        if [0, 16].include?(status_code)
          res = nil
          File.open(res_file, "r") {|f| res = JSON.load(f.read)}
          res
        else
          raise ::Dapp::Error::Base, code: :ruby2go_command_unexpected_exitstatus, data: { progname: progname, status_code: status_code }
        end
      end

      def _ruby2go_tmp_dir
        @_ruby2go_tmp_dir ||= Dir.mktmpdir('dapp-ruby2go-', tmp_base_dir)
      end

      def _download_ruby2go_bin(progname, bin_path)
        lock("downloader.bin.#{progname}", default_timeout: 1800) do
          return if File.exists? bin_path

          log_process("Downloading #{progname} dapp dependency") do
            location = URI("https://dl.bintray.com/flant/dapp/#{::Dapp::VERSION}/#{progname}")

            tmp_bin_path = File.join(self.class.tmp_base_dir, "#{progname}-#{SecureRandom.uuid}")
            ::Dapp::Downloader.download(location, tmp_bin_path, show_progress: true, progress_titile: bin_path)

            checksum_location = URI("https://dl.bintray.com/flant/dapp/#{::Dapp::VERSION}/#{progname}.sha")
            tmp_bin_checksum_path = tmp_bin_path + ".checksum"
            ::Dapp::Downloader.download(checksum_location, tmp_bin_checksum_path)

            if Digest::SHA256.hexdigest(File.read(tmp_bin_path)) != File.read(tmp_bin_checksum_path).strip
              raise ::Dapp::Error::Dapp, code: :ruby2go_download_failed_bad_checksum, data: {url: location.to_s, checksum_url: checksum_location.to_s, progname: progname}
            end

            File.chmod(0755, tmp_bin_path)
            FileUtils.mkdir_p File.dirname(bin_path)
            FileUtils.mv tmp_bin_path, bin_path
          end # log_process
        end # lock
      end
    end # Ruby2Go
  end # Dapp
end # Dapp
