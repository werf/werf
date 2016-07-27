module Dapp
  # Application
  class Application
    # Path
    module Path
      def home_path(*path)
        make_path(config._home_path, *path).expand_path
      end

      def tmp_path(*path)
        make_path(@tmp_path, *path).expand_path.tap { |p| FileUtils.mkdir_p p.parent }
      end

      def metadata_path(*path)
        make_path(@metadata_path, home_path.basename, *path).expand_path.tap { |p| FileUtils.mkdir_p p.parent }
      end

      def container_dapp_path(*path)
        make_path('/.dapp', *path)
      end

      def container_tmp_path(*path)
        container_dapp_path('tmp', *path)
      end

      private

      def make_path(base, *path)
        path.compact.map(&:to_s).inject(Pathname.new(base), &:+)
      end
    end # Path
  end # Application
end # Dapp
