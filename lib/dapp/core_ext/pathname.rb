module Dapp
  # CoreExt
  module CoreExt
    # Pathname
    module Pathname
      def subpath_of?(another_path)
        another_path_descends = []
        ::Pathname.new(another_path).cleanpath.descend {|d| another_path_descends << d}

        path_descends = []
        cleanpath.descend {|d| path_descends << d}

        (path_descends & another_path_descends) == another_path_descends and
          (path_descends - another_path_descends).any?
      end

      def subpath_of(another_path)
        return unless subpath_of? another_path
        cleanpath.to_s.partition(::Pathname.new(another_path).cleanpath.to_s + '/').last
      end
    end # Pathname
  end # CoreExt
end # Dapp

::Pathname.send(:include, ::Dapp::CoreExt::Pathname)
