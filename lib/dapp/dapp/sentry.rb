module Dapp
  class Dapp
    module Sentry
      def sentry_message(msg, **kwargs)
        return if not ensure_sentry_configured
        kwargs[:level] ||= "info"
        Raven.capture_message(msg, _sentry_params(**kwargs))
      end

      def sentry_exception(exception, **kwargs)
        return if not ensure_sentry_configured
        Raven.capture_exception(exception, _sentry_params(**kwargs))
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

      def _sentry_params(level: nil, tags: {}, extra: {}, user: {})
        {
          level: level,
          tags:  _sentry_tags_context.merge(tags),
          extra: _sentry_extra_context.merge(extra),
          user:  _sentry_user_context.merge(user),
        }
      end

      def _sentry_extra_context
        {
          "pwd" => Dir.pwd,
          "dapp-dir" => self.work_dir,
          "build-dir" => self.build_dir,
          "options" => self.options,
          "DAPP_FORCE_SAVE_CACHE" => ENV["DAPP_FORCE_SAVE_CACHE"],
          "DAPP_BIN_DAPPFILE_YML" => ENV["DAPP_BIN_DAPPFILE_YML"],
          "ANSIBLE_ARGS" => ENV["ANSIBLE_ARGS"],
          "DAPP_CHEF_DEBUG" => ENV["DAPP_CHEF_DEBUG"],
        }.tap {|extra|
          if git_own_repo_exist?
            extra["git"] = {
              "remote_origin_url" => git_own_repo.remote_origin_url, # may contain https token
              "name" => self.git_url_to_name(git_own_repo.remote_origin_url),
              "path" => git_own_repo.path,
              "workdir_path" => git_own_repo.workdir_path,
              "latest_commit" => git_own_repo.latest_commit,
            }
          end
        }
      end

      def _sentry_tags_context
        {
          "dapp-name" => self.name,
        }.tap {|tags|
          if git_own_repo_exist?
            tags["git-host"] = self.get_host_from_git_url(git_own_repo.remote_origin_url)

            git_name = self.git_url_to_name(git_own_repo.remote_origin_url)

            tags["git-group"] = git_name.partition("/")[0]
            tags["git-name"] = git_name
          end
        }
      end

      def _sentry_user_context
        {}
      end
    end # Sentry
  end # Dapp
end # Dapp
