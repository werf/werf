module Dapp
  # Application
  class Application
    # Path
    module Path
      def home_path(*path)
        make_path(config._home_path, *path).expand_path
      end

      def build_path(*path)
        make_path(@build_path, *path).expand_path.tap { |p| FileUtils.mkdir_p p.parent }
      end

      def build_cache_path(*path)
        make_path(@build_cache_path, *path).expand_path.tap { |p| FileUtils.mkdir_p p.parent }
      end

      def container_build_path(*path)
        make_path('/.build', *path)
      end

      private

      def make_path(base, *path)
        path.compact.map(&:to_s).inject(Pathname.new(base), &:+)
      end
    end # Path
  end # Application
end # Dapp
