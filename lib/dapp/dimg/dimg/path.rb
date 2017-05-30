module Dapp
  module Dimg
    class Dimg
      module Path
        def home_path(*path)
          make_path(dapp.path, *path).expand_path
        end

        def tmp_base_dir
          File.expand_path(dapp.options[:tmp_dir_prefix] || '/tmp')
        end

        def tmp_path(*path)
          @tmp_path ||= Dir.mktmpdir('dapp-', tmp_base_dir)
          make_path(@tmp_path, *path).expand_path.tap { |p| p.parent.mkpath }
        end

        def build_path(*path)
          make_path(dapp.build_path, *path).expand_path.tap { |p| p.parent.mkpath }
        end

        def container_dapp_path(*path)
          make_path('/.dapp', *path)
        end

        def container_tmp_path(*path)
          container_dapp_path('tmp', *path)
        end

        alias build_dir build_path
        alias tmp_dir tmp_path

        private

        def make_path(base, *path)
          path.compact.map(&:to_s).inject(Pathname.new(base), &:+)
        end
      end # Path
    end
  end # Dimg
end # Dapp
