module Dapp
  module Dimg
    # Image
    module Image
      # Stage
      class Scratch < Stage
        def initialize(**_kwargs)
          super
          @from_archives = []
        end

        def add_archive(*archives)
          @from_archives.concat(archives.flatten)
        end

        def build!(**_kwargs)
          raise

          # FIXME: system-shellout rejected
          # build_from_command = if from_archives.empty?
          #                        "#{dapp.tar_bin} c --files-from /dev/null"
          #                      else
          #                        "#{dapp.cat_bin} #{from_archives.join(' ')}"
          #                      end
          # @built_id = dapp.system_shellout!("#{build_from_command} | docker import #{prepared_change} - ").stdout.strip
        end

        protected

        attr_accessor :from_archives
      end # Stage
    end # Image
  end # Dimg
end # Dapp
