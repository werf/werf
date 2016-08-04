module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Artifact
        module Artifact
          def before_artifacts
            @before_artifacts ||= do_artifacts(application.config._artifact.select { |artifact| artifact._before == name })
          end

          def after_artifacts
            @after_artifacts ||= do_artifacts(application.config._artifact.select { |artifact| artifact._after == name })
          end

          def do_artifacts(artifacts)
            verbose = application.log_verbose?
            artifacts.map do |artifact|
              process = application.t(code: 'process.artifact_building', data: { name: artifact._config._name })
              application.log_secondary_process(process, short: !verbose) do
                application.with_log_indent do
                  {
                    name: artifact._config._name,
                    options: artifact._artifact_options,
                    app: Application.new(artifact_app_options(artifact, verbose)).tap(&:build!)
                  }
                end
              end
            end
          end

          def artifact_app_options(artifact, verbose)
            {
              config: artifact._config,
              cli_options: application.cli_options.merge(log_quiet: !verbose),
              ignore_git_fetch: application.ignore_git_fetch,
              is_artifact: true
            }
          end

          def artifacts_signatures
            (before_artifacts + after_artifacts).map { |artifact| hashsum [artifact[:app].signature, artifact[:options]] }
          end

          # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
          def apply_artifact(artifact, image)
            return if application.dry_run?

            artifact_name = artifact[:name]
            app = artifact[:app]
            cwd = artifact[:options][:cwd]
            paths = artifact[:options][:paths]
            owner = artifact[:options][:owner]
            group = artifact[:options][:group]
            where_to_add = artifact[:options][:where_to_add]

            docker_options = ['--rm',
                              "--volume #{application.tmp_path('artifact', artifact_name)}:#{app.container_tmp_path(artifact_name)}",
                              '--entrypoint /bin/sh']
            commands = safe_cp(where_to_add, app.container_tmp_path(artifact_name), Process.uid, Process.gid)
            application.log_secondary_process(application.t(code: 'process.artifact_copy', data: { name: artifact_name }), short: true) do
              app.run(docker_options, Array(application.shellout_pack(commands)))
            end

            commands = safe_cp(application.container_tmp_path('artifact', artifact_name), where_to_add, owner, group, cwd, paths)
            image.add_commands commands
          end
          # rubocop:enable Metrics/AbcSize, Metrics/MethodLength

          # rubocop:disable Metrics/ParameterLists
          def safe_cp(from, to, owner, group, cwd = '', paths = [])
            credentials = ''
            credentials += "-o #{owner} " if owner
            credentials += "-g #{group} " if group

            copy_files = lambda do |from_, cwd_, path_ = ''|
              "find #{File.join(from_, cwd_, path_)} -type f -exec bash -ec 'install -D #{credentials} {} " \
            "#{File.join(to, "$(echo {} | sed -e \"s/#{File.join(from_, cwd_).gsub('/', '\\/')}//g\")")}' \\;"
            end

            commands = []
            commands << ['install', credentials, '-d', to].join(' ')
            commands.concat(paths.empty? ? Array(copy_files.call(from, cwd)) : paths.map { |path| copy_files.call(from, cwd, path) })
            commands << "find #{to} -type d -exec bash -ec 'install -d #{credentials} {}' \\;"
            commands.join(' && ')
          end
          # rubocop:enable Metrics/ParameterLists
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
