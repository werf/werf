module Dapp
  # Project
  class Project
    # Command
    module Command
      # StagesFlush
      module StagesFlush
        def stages_flush
          build_configs.map(&:_basename).uniq.each do |basename|
            log(basename)
            containers_flush(basename)
            remove_images(%(docker images --format="{{.Repository}}:{{.Tag}}" #{stage_cache(basename)}))
          end
        end
      end
    end
  end # Project
end # Dapp
