module Dapp
  module Config
    # GitArtifact
    class GitArtifact
      attr_reader :_local
      attr_reader :_remote

      def initialize
        @_local  = []
        @_remote = []
      end

      def local(*args)
        @_local.tap { |local| local << Local.new(*args) unless args.empty? }
      end

      def remote(*args)
        @_remote.tap { |remote| remote << Remote.new(*args) unless args.empty? }
      end

      protected

      def clone
        Marshal.load(Marshal.dump(self))
      end

      # Local
      class Local < Artifact::Base
        protected

        def code
          :git_artifact_unexpected_attribute
        end
      end

      # Remote
      class Remote < Local
        attr_accessor :_url, :_name, :_branch

        def initialize(url, where_to_add, **options)
          @_url          = url
          @_name         = url.gsub(%r{.*?([^\/ ]+)\.git}, '\\1')
          @_branch       = options.delete(:branch)
          super(where_to_add, **options)
        end

        def _artifact_options
          super.merge(name: _name, branch: _branch)
        end
      end
    end
  end
end
