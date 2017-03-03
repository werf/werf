module Dapp
  module Dimg
    class Dapp
      module Command
        module BuildContext
          module Common
            def build_context_images_tar
              build_context_path('images.tar')
            end

            def build_context_build_tar
              build_context_path('build.tar')
            end

            def build_context_path(*path)
              path.compact.map(&:to_s).inject(Pathname.new(build_context_directory), &:+).tap { |p| p.parent.mkpath }
            end

            def build_context_directory
              File.expand_path(cli_options[:build_context_directory].to_s)
            end
          end
        end
      end
    end
  end
end
