module Dapp
  # FIXME Move somewhere inside builder/chef
  class Berksfile
    class Parser
      def initialize(berksfile)
        @berksfile = berksfile
        parse
      end

      def cookbook(name, path: nil, **kwargs)
        @berksfile.add_local_cookbook_path(path) if path
      end

      def method_missing(*a, &blk)
      end

      private

      def parse
        instance_eval(@berksfile.path.read, @berksfile.path.to_s)
      end
    end # Parser

    attr_reader :application
    attr_reader :path
    attr_reader :local_cookbook_paths

    def initialize(application, path)
      @application = application
      @path = path
      @local_cookbook_paths = []
      @parser = Parser.new(self)
    end

    def add_local_cookbook_path(path)
      @local_cookbook_paths << application.home_path(path)
    end
  end # Berksfile
end # Dapp
