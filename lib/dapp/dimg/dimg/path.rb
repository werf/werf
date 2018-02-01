module Dapp
  module Dimg
    class Dimg
      module Path
        def home_path(*path)
          dapp.path(*path).expand_path
        end

        def tmp_dir_exists?
          @tmp_path != nil
        end

        def tmp_path(*path)
          @tmp_path ||= Dir.mktmpdir('dapp-', dapp.tmp_base_dir)
          make_path(@tmp_path, *path).expand_path.tap { |p| p.parent.mkpath }
        end

        def build_path(*path)
          dapp.build_path(*path).expand_path.tap { |p| p.parent.mkpath }
        end

        def container_dapp_path(*path)
          make_path('/.dapp', *path)
        end

        def container_tmp_path(*path)
          container_dapp_path('tmp', *path)
        end

        alias build_dir build_path
        alias tmp_dir tmp_path
      end # Path
    end
  end # Dimg
end # Dapp
