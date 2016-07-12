module Dapp
  module Config
    class GitArtifact < Base
      attr_reader :_local
      attr_reader :_remote

      def initialize
        @_local  = []
        @_remote = []
        super
      end

      def local(*args)
        @_local.tap { |local| local << Local.new(*args) unless args.empty? }
      end

      def remote(*args)
        @_remote.tap { |remote| remote << Remote.new(*args) unless args.empty? }
      end

      def to_h
        { local: _local.map(&:to_h), remote: _remote.map(&:to_h) }
      end

      class Local < Base
        attr_accessor :_where_to_add, :_cwd, :_paths, :_owner, :_group

        def initialize(where_to_add, **options)
          @_cwd          = '/'
          @_where_to_add = where_to_add

          options.each do |k, v|
            if respond_to? "_#{k}="
              send(:"_#{k}=", v)
            else
              raise "'#{object_name}' git artifact doesn't have attribute '#{k}'!"
            end
          end
          super()
        end

        def artifact_options
          to_h
        end

        def to_h
          {
            where_to_add: _where_to_add,
            cwd:          _cwd,
            paths:        _paths,
            owner:        _owner,
            group:        _group
          }
        end

        protected

        def object_name
          self.class.to_s.split('::').last
        end
      end

      class Remote < Local
        attr_accessor :_url, :_name, :_branch, :_ssh_key_path

        def initialize(url, where_to_add, **options)
          @_url          = url
          @_name         = url.gsub(%r{.*?([^\/ ]+)\.git}, '\\1')
          @_branch       = options.delete(:branch)
          @_ssg_key_path = options.delete(:ssg_key_path)
          super(where_to_add, **options)
        end

        def artifact_options
          super.merge({ name: _name, branch: _branch })
        end
      end
    end
  end
end
