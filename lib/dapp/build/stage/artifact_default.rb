module Dapp
  module Build
    module Stage
      # ArtifactDefault
      class ArtifactDefault < ArtifactBase
        protected

        # rubocop:disable Metrics/AbcSize
        def apply_artifact(artifact, image)
          return if dimg.project.dry_run?

          artifact_name = artifact[:name]
          artifact_dimg = artifact[:dimg]
          cwd = artifact[:options][:cwd]
          paths = artifact[:options][:paths]
          exclude_paths = artifact[:options][:exclude_paths]
          owner = artifact[:options][:owner]
          group = artifact[:options][:group]
          where_to_add = artifact[:options][:where_to_add]

          command = safe_cp(where_to_add, artifact_dimg.container_tmp_path(artifact_name), Process.uid, Process.gid, cwd, paths, exclude_paths)
          run_artifact_dimg(artifact_dimg, artifact_name, command)

          command = safe_cp(dimg.container_tmp_path('artifact', artifact_name), where_to_add, owner, group, '', paths, exclude_paths)
          image.add_command command
          image.add_volume "#{dimg.tmp_path('artifact', artifact_name)}:#{dimg.container_tmp_path('artifact', artifact_name)}:ro"
        end
        # rubocop:enable Metrics/AbcSize, Metrics/MethodLength

        private

        # rubocop:disable Metrics/ParameterLists, Metrics/AbcSize, Metrics/MethodLength
        def safe_cp(from, to, owner, group, cwd = '', paths = [], exclude_paths = [])
          credentials = ''
          credentials += "-o #{owner} " if owner
          credentials += "-g #{group} " if group
          excludes = find_command_excludes(from, cwd, exclude_paths).join(' ')

          copy_files = proc do |from_, cwd_, path_ = ''|
            cwd_ = File.expand_path(File.join('/', cwd_))
            "#{dimg.project.find_path} #{File.join(from_, cwd_, path_)} #{excludes} -type f -exec " \
            "#{dimg.project.bash_path} -ec '#{dimg.project.install_path} -D #{credentials} {} " \
            "#{File.join(to, "$(echo {} | " \
            "#{dimg.project.sed_path} -e \"s/#{File.join(from_, cwd_).gsub('/', '\\/')}//g\")")}' \\;"
          end

          commands = []
          commands << [dimg.project.install_path, credentials, '-d', to].join(' ')
          commands.concat(paths.empty? ? Array(copy_files.call(from, cwd)) : paths.map { |path| copy_files.call(from, cwd, path) })
          commands << "#{dimg.project.find_path} #{to} -type d -exec " \
                      "#{dimg.project.bash_path} -ec '#{dimg.project.install_path} -d #{credentials} {}' \\;"
          commands.join(' && ')
        end
        # rubocop:enable Metrics/ParameterLists, Metrics/AbcSize, Metrics/MethodLength

        def find_command_excludes(from, cwd, exclude_paths)
          exclude_paths.map { |path| "-not \\( -path #{File.join(from, cwd, path)} -prune \\)" }
        end
      end # ArtifactDefault
    end # Stage
  end # Build
end # Dapp
