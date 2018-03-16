module Dapp
  class Dapp
    module Sentry
      def sentry_message(msg, **kwargs)
        return if not ensure_sentry_configured
        kwargs[:level] ||= "info"
        Raven.capture_message(msg, _make_sentry_params(**kwargs))
      end

      def sentry_exception(exception, **kwargs)
        return if not ensure_sentry_configured
        (kwargs[:tags] ||= {})['error-code'] = begin
          net_status = exception.net_status
          [net_status[:context], net_status[:code]].compact.join('_')
        end
        Raven.capture_exception(exception, _make_sentry_params(**kwargs))
      end

      def ensure_sentry_configured
        return false unless sentry_settings = settings["sentry"]

        unless @sentry_settings_configured
          Raven.configure do |config|
            logger = ::Logger.new(STDOUT)
            logger.level = ::Logger::WARN

            config.logger = logger
            config.dsn = sentry_settings["dsn"]
          end

          @sentry_settings_configured = true
        end

        return true
      end

      def _make_sentry_params(level: nil, tags: {}, extra: {}, user: {})
        {
          level: level,
          tags:  _sentry_tags_context.merge(tags),
          extra: _sentry_extra_context.merge(extra),
          user:  _sentry_user_context.merge(user),
        }
      end

      def _sentry_extra_context
        @_sentry_extra_context ||= {
          "pwd" => Dir.pwd,
          "dapp-dir" => self.work_dir,
          "options" => self.options,
          "env-options" => {
            "DAPP_FORCE_SAVE_CACHE" => ENV["DAPP_FORCE_SAVE_CACHE"],
            "DAPP_BIN_DAPPFILE_YML" => ENV["DAPP_BIN_DAPPFILE_YML"],
            "ANSIBLE_ARGS" => ENV["ANSIBLE_ARGS"],
            "DAPP_CHEF_DEBUG" => ENV["DAPP_CHEF_DEBUG"],
          },
        }.tap {|extra|
          extra["ci-env"] = {"CI" => ENV["CI"]}
          ENV.select {|k, v| k.start_with?("CI_")}.each do |k, v|
            extra["ci-env"][k] = v
          end
        }
      end

      def _sentry_tags_context
        name = options[:name] || 
        @_sentry_tags_context ||= {
          "dapp-short-version" => ::Dapp::VERSION.split(".")[0..1].join("."),
          "dapp-version" => ::Dapp::VERSION,
          "dapp-build-cache-version" => ::Dapp::BUILD_CACHE_VERSION,
          "dapp-command" => self.options[:dapp_command],
        }.tap {|tags|
          git_config_path = File.join(Dir.pwd, ".git/config")

          tags["dapp-name"] = options[:name]

          if File.exists? git_config_path
            cfg = IniFile.load(File.join(Dir.pwd, ".git/config"))
            remote_origin_cfg = cfg['remote "origin"']
            remote_origin_url = remote_origin_cfg["url"]
            if remote_origin_url
              tags["dapp-name"] ||= begin
                repo_name = remote_origin_url.split('/').last
                repo_name = repo_name[/.*(?=\.git)/] if repo_name.end_with? '.git'
                repo_name
              end

              tags["git-host"] = self.get_host_from_git_url(remote_origin_url)

              git_name = self.git_url_to_name(remote_origin_url)

              tags["git-group"] = git_name.partition("/")[0]
              tags["git-name"] = git_name
            end
          end

          tags["dapp-name"] ||= File.basename(Dir.pwd)

          begin
            ver = self.class.host_docker_minor_version
            tags["docker-minor-version"] = ver.to_s
          rescue ::Exception
          end
        }
      end

      def _sentry_user_context
        @_sentry_user_context ||= {}
      end
    end # Sentry
  end # Dapp
end # Dapp
