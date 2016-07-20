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

      def clone
        Marshal.load(Marshal.dump(self))
      end

      # Local
      class Local
        attr_accessor :_where_to_add, :_cwd, :_paths, :_owner, :_group

        def initialize(where_to_add, **options)
          @_cwd          = ''
          @_where_to_add = where_to_add

          options.each do |k, v|
            respond_to?("_#{k}=") ? send(:"_#{k}=", v) : fail(Error::Config, code: :git_artifact_unexpected_attribute,
                                                                             data: { type: object_name, attr: k })
          end
        end

        def _artifact_options
          {
            where_to_add: _where_to_add,
            cwd:          _cwd,
            paths:        _paths,
            owner:        _owner,
            group:        _group
          }
        end

        def clone
          Marshal.load(Marshal.dump(self))
        end

        protected

        def object_name
          self.class.to_s.split('::').last
        end
      end

      # Remote
      class Remote < Local
        attr_accessor :_url, :_name, :_branch, :_ssh_key_path

        def initialize(url, where_to_add, **options)
          @_url          = url
          @_name         = url.gsub(%r{.*?([^\/ ]+)\.git}, '\\1')
          @_branch       = options.delete(:branch)
          @_ssg_key_path = options.delete(:ssg_key_path)
          super(where_to_add, **options)
        end

        def _artifact_options
          super.merge(name: _name, branch: _branch)
        end
      end
    end
  end
end
