module Dapp
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
        build_from_command = if from_archives.empty?
                               "#{project.tar_path} c --files-from /dev/null"
                             else
                               "#{project.cat_path} #{from_archives.join(' ')}"
                             end
        @built_id = project.system_shellout!("#{build_from_command} | docker import #{prepared_change} - ").stdout.strip
      end

      protected

      attr_accessor :from_archives
    end # Stage
  end # Image
end # Dapp
