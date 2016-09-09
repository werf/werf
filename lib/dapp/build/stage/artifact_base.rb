module Dapp
  module Build
    module Stage
      # ArtifactBase
      class ArtifactBase < Base
        def dependencies
          artifacts_signatures
        end

        def prepare_image
          super
          artifacts_applications_build!
          artifacts_labels = {}
          artifacts.each do |artifact|
            apply_artifact(artifact, image)
            artifacts_labels["dapp-artifact-#{artifact[:name]}".to_sym] = artifact[:app].send(:last_stage).image.built_id
          end
          image.add_service_change_label artifacts_labels
        end

        def images
          [image].concat(artifacts.map { |artifact| artifact[:app].images }.flatten)
        end

        protected

        def should_not_be_detailed?
          true
        end

        def ignore_log_commands?
          true
        end

        def artifacts
          @artifacts ||= begin
            application.config.public_send("_#{name}").map do |artifact|
              app = Dapp::Artifact.new(config: artifact._config,
                                       project: application.project,
                                       ignore_git_fetch: application.ignore_git_fetch)
              { name: artifact._config._name, options: artifact._artifact_options, app: app }
            end
          end
        end

        def artifacts_signatures
          artifacts.map { |artifact| hashsum [artifact[:app].signature, artifact[:options]] }
        end

        def artifacts_applications_build!
          artifacts.each do |artifact|
            process = application.project.t(code: 'process.artifact_building', data: { name: artifact[:name] })
            application.project.log_secondary_process(process,
                                                      short: !application.project.log_verbose?,
                                                      quiet: application.artifact? && !application.project.log_verbose?) do
              application.project.with_log_indent do
                artifact[:app].build!
              end
            end
          end
        end

        # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
        def apply_artifact(artifact, image)
          return if application.project.dry_run?

          artifact_name = artifact[:name]
          app = artifact[:app]
          cwd = artifact[:options][:cwd]
          paths = artifact[:options][:paths]
          exclude_paths = artifact[:options][:exclude_paths]
          owner = artifact[:options][:owner]
          group = artifact[:options][:group]
          where_to_add = artifact[:options][:where_to_add]

          docker_options = ['--rm',
                            "--volume #{application.tmp_path('artifact', artifact_name)}:#{app.container_tmp_path(artifact_name)}",
                            "--volumes-from #{application.project.base_container}",
                            "--entrypoint #{application.project.bash_path}"]
          commands = safe_cp(where_to_add, app.container_tmp_path(artifact_name), Process.uid, Process.gid)
          application.project.log_secondary_process(application.project.t(code: 'process.artifact_copy',
                                                                          data: { name: artifact_name }),
                                                    short: true,
                                                    quiet: application.artifact? && !application.project.log_verbose?) do
            app.run(docker_options, [%(-ec '#{application.project.shellout_pack(commands)}')])
          end

          commands = safe_cp(application.container_tmp_path('artifact', artifact_name), where_to_add, owner, group, cwd, paths, exclude_paths)
          image.add_command commands
          image.add_volume "#{application.tmp_path('artifact', artifact_name)}:#{application.container_tmp_path('artifact', artifact_name)}:ro"
        end
        # rubocop:enable Metrics/AbcSize, Metrics/MethodLength

        private

        # rubocop:disable Metrics/ParameterLists
        def safe_cp(from, to, owner, group, cwd = '', paths = [], exclude_paths = [])
          credentials = ''
          credentials += "-o #{owner} " if owner
          credentials += "-g #{group} " if group
          excludes = find_command_excludes(from, cwd, exclude_paths).join(' ')

          copy_files = proc do |from_, cwd_, path_ = ''|
            cwd_ = File.expand_path(File.join('/', cwd_))
            "find #{File.join(from_, cwd_, path_)} #{excludes} -type f " \
            "-exec #{application.project.bash_path} -ec 'install -D #{credentials} {} " \
            "#{File.join(to, "$(echo {} | sed -e \"s/#{File.join(from_, cwd_).gsub('/', '\\/')}//g\")")}' \\;"
          end

          commands = []
          commands << ['install', credentials, '-d', to].join(' ')
          commands.concat(paths.empty? ? Array(copy_files.call(from, cwd)) : paths.map { |path| copy_files.call(from, cwd, path) })
          commands << "find #{to} -type d -exec #{application.project.bash_path} -ec 'install -d #{credentials} {}' \\;"
          commands.join(' && ')
        end
        # rubocop:enable Metrics/ParameterLists

        def find_command_excludes(from, cwd, exclude_paths)
          exclude_paths.map { |path| "-not \\( -path #{File.join(from, cwd, path)} -prune \\)" }
        end
      end # ArtifactBase
    end # Stage
  end # Build
end # Dapp
