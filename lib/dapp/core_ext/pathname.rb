module Dapp
  module CoreExt
    module Pathname
      def subpath_of?(another_path)
        another_path_descends = []
        ::Pathname.new(another_path).cleanpath.descend {|d| another_path_descends << d}

        path_descends = []
        cleanpath.descend {|d| path_descends << d}

        (path_descends & another_path_descends) == another_path_descends
      end

      def subpath_of(another_path)
        another_cleanpath = ::Pathname.new(another_path).cleanpath

        return     unless subpath_of? another_path
        return '.' if cleanpath.to_s == another_cleanpath.to_s
        cleanpath.to_s.partition(another_cleanpath.to_s + '/').last
      end
    end # Pathname
  end # CoreExt
end # Dapp

::Pathname.send(:include, ::Dapp::CoreExt::Pathname)
