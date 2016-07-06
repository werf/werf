module Dapp
  module Config
    class GitArtifact < Base
      def initialize(main_conf, &blk)
        @remote = []
        super
      end

      def local(*args)
        # FIXME
        @local ||= Local.new(main_conf, *args)
      end

      def remote(*args)
        @remote.tap do |remote|
          remote << Remote.new(main_conf, *args) unless args.empty?
        end
      end

      class Local < Base
        attr_accessor :where_to_add, :cwd, :paths, :owner, :group

        def initialize(main_conf, where_to_add, **options, &blk)
          @cwd ||= '/'
          @where_to_add = where_to_add
          super(main_conf, **options, &blk)
        end
      end

      class Remote < Local
        attr_accessor :name, :branch, :ssh_key_path

        def initialize(main_conf, url, where_to_add, **options, &blk)
          @name = url.gsub(%r{.*?([^\/ ]+)\.git}, '\\1')
          @branch = options.delete(:branch) || shellout!("git -C #{main_conf.home_path} rev-parse --abbrev-ref HEAD").stdout.strip
          @ssg_key_path = options.delete(:ssg_key_path)
          super(main_conf, where_to_add, **options, &blk)
        end
      end
    end
  end
end
