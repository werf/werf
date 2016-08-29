module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        module FlushLocal
          def stages_flush_local
            build_configs.map(&:_basename).uniq.each do |basename|
              lock("#{basename}.images") do
                log(basename)
                containers_flush(basename)
                remove_images(%(docker images --format="{{.Repository}}:{{.Tag}}" #{stage_cache(basename)}))
              end
            end
          end
        end
      end
    end
  end # Project
end # Dapp
