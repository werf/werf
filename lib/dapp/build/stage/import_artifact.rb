module Dapp
  module Build
    module Stage
      # ImportArtifact
      class ImportArtifact < ArtifactBase
        def initialize(dimg)
          @dimg = dimg
        end

        def signature
          hashsum [*dependencies.flatten, change_options]
        end

        def image
          @image ||= Image::Scratch.new(name: image_name, project: dimg.project)
        end

        def image_add_volumes
        end

        def prepare_image
          super
          change_options.each do |k, v|
            image.public_send("add_change_#{k}", v)
          end
        end

        protected

        # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
        def apply_artifact(artifact, image)
          return if dimg.project.dry_run?

          artifact_name = artifact[:name]
          artifact_dimg = artifact[:dimg]
          cwd = artifact[:options][:cwd]
          include_paths = artifact[:options][:include_paths]
          owner = artifact[:options][:owner]
          group = artifact[:options][:group]
          to = artifact[:options][:to]

          sudo = dimg.project.sudo_command(owner: Process.uid, group: Process.gid)

          credentials = ''
          credentials += "--owner=#{owner} " if owner
          credentials += "--group=#{group} " if group
          credentials += '--numeric-owner'

          archive_path = dimg.tmp_path('artifact', artifact_name, 'archive.tar.gz')
          container_archive_path = File.join(artifact_dimg.container_tmp_path(artifact_name), 'archive.tar.gz')

          exclude_paths = artifact[:options][:exclude_paths].map { |path| "--exclude=#{path}" }.join(' ')
          include_paths = if include_paths.empty?
                    [File.join(to, cwd, '*')]
                          else
                    include_paths.map { |path| File.join(to, cwd, path, '*') }
                          end
          include_paths.map! { |path| path[1..-1] } # relative path

          command = "#{sudo} #{dimg.project.tar_path} -czf #{container_archive_path} #{exclude_paths} #{include_paths.join(' ')} #{credentials}"
          run_artifact_dimg(artifact_dimg, artifact_name, command)

          image.add_archive archive_path
        end
        # rubocop:enable Metrics/AbcSize, Metrics/MethodLength
      end # ImportArtifact
    end # Stage
  end # Build
end # Dapp
