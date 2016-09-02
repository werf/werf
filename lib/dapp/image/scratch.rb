module Dapp
  # Image
  module Image
    # Stage
    class Scratch < Stage
      def initialize(project:)
        @project = project
        build!
        super(name: 'dappdeps/scratch', project: project, built_id: built_id, from: nil)
      end

      def build!
        return if project.dry_run?
        @built_id = project.shellout!('tar c --files-from /dev/null | docker import - dappdeps/scratch').stdout.strip
      end

      def pull!(*_args)
      end

      def pulled?
        false
      end
    end # Stage
  end # Image
end # Dapp
