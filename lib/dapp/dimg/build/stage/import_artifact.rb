module Dapp
  module Dimg
    module Build
      module Stage
        class ImportArtifact < ArtifactBase
          def initialize(dimg)
            @dimg = dimg
          end

          def signature
            hashsum [*dependencies.flatten, change_options]
          end

          def image
            @image ||= Image::Scratch.new(name: image_name, dapp: dimg.dapp)
          end

          def image_add_mounts
          end

          def prepare_image
            super do
              change_options.each do |k, v|
                image.public_send("add_change_#{k}", v)
              end
            end
          end

          protected

          def apply_artifact(artifact, image)
            return if dimg.dapp.dry_run?

            artifact_dimg = artifact[:dimg]
            cwd = artifact[:options][:cwd]
            to = artifact[:options][:to]
            include_paths = artifact[:options][:include_paths]
            owner = artifact[:options][:owner]
            group = artifact[:options][:group]

            sudo = dimg.dapp.sudo_command(owner: Process.uid, group: Process.gid)

            credentials = ''
            credentials += "--owner=#{owner} " if owner
            credentials += "--group=#{group} " if group
            credentials += '--numeric-owner'

            archive_path = dimg.tmp_path('artifact', artifact_dimg.name, 'archive.tar.gz')
            container_archive_path = File.join(artifact_dimg.container_tmp_path(artifact_dimg.name), 'archive.tar.gz')

            exclude_paths = artifact[:options][:exclude_paths].map { |path| "--exclude=#{path}" }.join(' ')
            include_paths = include_paths.empty? ? [File.join(cwd, '*')] : include_paths.map { |path| File.join(cwd, path, '*') }
            include_paths.map! { |path| path[1..-1] } # relative path

            command = "#{sudo} #{dimg.dapp.tar_bin} #{tar_option_transform(cwd, to)} -czf #{container_archive_path} #{exclude_paths} #{include_paths.join(' ')} #{credentials}"
            run_artifact_dimg(artifact_dimg, command)

            image.add_archive archive_path
          end

          private

          def tar_option_transform(cwd, to)
            format = proc do |path|
              res = path.chomp('/').reverse.chomp('/').reverse
              res = res.gsub('/', '\/')
              res
            end
            "--transform \"s/^#{format.call(cwd)}/#{format.call(to)}/\""
          end
        end # ImportArtifact
      end # Stage
    end # Build
  end # Dimg
end # Dapp
