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

      def ruby2go_init
        @_call_before_terminate << proc {
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

        cmd_res = shellout! "#{bin_path} -args-from-file #{args_file} -result-to-file #{res_file}", raise_on_error: false, verbose: true
        if [0, 16].include? cmd_res.exitstatus
          res = nil
          File.open(res_file, "r") {|f| res = JSON.load(f.read)}
          return res
        else
          shellout_cmd_should_succeed! cmd_res
        end
      end

      def _ruby2go_tmp_dir
        @_ruby2go_tmp_dir ||= Dir.mktmpdir('dapp-ruby2go-', tmp_base_dir)
      end

      def _download_ruby2go_bin(progname, bin_path)
        # TODO
      end
    end # Ruby2Go
  end # Dapp
end # Dapp
