module Dapp
  module Build
    module Stage
      # ImportArtifact
      class ImportArtifact < ArtifactBase
        def signature
          hashsum [*dependencies.flatten]
        end

        def image
          @image ||= Image::Scratch.new(name: image_name, project: application.project)
        end

        def image_add_tmp_volumes(_type)
        end

        protected

        # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
        def apply_artifact(artifact, image)
          return if application.project.dry_run?

          artifact_name = artifact[:name]
          app = artifact[:app]
          cwd = artifact[:options][:cwd]
          paths = artifact[:options][:paths]
          owner = artifact[:options][:owner]
          group = artifact[:options][:group]
          where_to_add = artifact[:options][:where_to_add]

          sudo = application.project.sudo_command(owner: Process.uid, group: Process.gid)

          credentials = ''
          credentials += "--owner=#{owner} " if owner
          credentials += "--group=#{group} " if group
          credentials += '--numeric-owner'

          archive_path = application.tmp_path('artifact', artifact_name, 'archive.tar.gz')
          container_archive_path = File.join(app.container_tmp_path(artifact_name), 'archive.tar.gz')

          exclude_paths = artifact[:options][:exclude_paths].map { |path| "--exclude=#{path}" }.join(' ')
          paths = if paths.empty?
                    [File.join(where_to_add, cwd, '*')]
                  else
                    paths.map { |path| File.join(where_to_add, cwd, path, '*') }
                  end
          paths.map! { |path| path[1..-1] } # relative path

          command = "#{sudo} #{application.project.tar_path} -czf #{container_archive_path} #{exclude_paths} #{paths.join(' ')} #{credentials}"
          run_artifact_app(app, artifact_name, command)

          image.add_archive archive_path
        end
        # rubocop:enable Metrics/AbcSize, Metrics/MethodLength
      end # ImportArtifact
    end # Stage
  end # Build
end # Dapp
