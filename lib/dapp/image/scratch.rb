module Dapp
  # Image
  module Image
    # Stage
    class Scratch < Stage
      def initialize(project:)
        @name = 'dappdeps/scratch:latest'
        @project = project
        id || build!
        super(name: name, project: project)
      end

      def build!
        return if project.dry_run?
        project.shellout!("#{project.tar_path} c --files-from /dev/null | docker import - dappdeps/scratch").stdout.strip
        cache_reset
      end

      def pull!(*_args)
      end
    end # Stage
  end # Image
end # Dapp
