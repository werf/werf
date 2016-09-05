module Dapp
  # Project
  class Project
    # Command
    module Command
      # Cleanup
      module Cleanup
        def cleanup
          build_configs.map(&:_basename).uniq.each do |basename|
            lock("#{basename}.images") do
              log_step_with_indent(basename) do
                project_containers_flush(basename)
                project_dangling_images_flush(basename)
                remove_images_by_query([
                  'docker images',
                  %(--format '{{if ne "#{stage_cache(basename)}" .Repository }}{{.ID}}{{ end }}'),
                  %(-f "label=dapp=#{stage_dapp_label(basename)}")
                ].join(' ')) # FIXME: negative filter is not currently supported by the Docker CLI
              end
            end
          end
        end
      end
    end
  end # Project
end # Dapp
