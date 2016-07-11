module Dapp
  module Config
    # FIXME rename Application
    class Main < Base
      def initialize(**options)
        @attrs = options

        @attrs[:home_path]     ||= Pathname.new(@attrs[:dappfile_path]).parent.expand_path.to_s
        @attrs[:name]          ||= Pathname.new(@attrs[:home_path]).basename unless @attrs[:name]

        # FIXME Chef.detect ? :chef : :shell
        # FIXME 

        @attrs[:builder]       ||= :shell
        @attrs[:cache_version] ||= {}
        @apps                  = []

        super()
      end

      # FIXME remove, add 4 methods: chef, shell, git_artifact, docker
      def method_missing(name, *args)
        return attrs[name] if attrs.key?(name)
        klass       = Config.const_get(name.to_s.split('_').map(&:capitalize).join)
        attrs[name] ||= klass.new(self, *args)
      rescue NameError
        super
      end

      def builder_validation(builder_name)
        raise RuntimeError, "Builder type '#{builder_name}' is not defined!" unless attrs[:builder] == builder_name
      end

      def name(*args)
        option(:name, *args)
      end

      def builder(*args)
        # FIXME remove other builder object
        option(:builder, *args)
      end

      def apps
        @apps.empty? ? [self] : @apps.flatten
      end

      def cache_key(key = nil)
        @attrs[:cache_version][key]
      end

      def to_h
        hash = { name: name, builder: builder}
        hash.merge!(attrs.keys.inject({}) do |total, key|
          value = attrs[key]
          total[key] = value.to_h if !value.nil? && value.is_a?(Base)
          total
        end)
        hash[:cache_key] = attrs[:cache_version] unless attrs[:cache_version].empty?
        hash
      end

      private

      def attrs
        @attrs ||= {}
      end

      def cache_version(value = nil, **options)
        if value.nil?
          options.each { |k, v| @attrs[:cache_version][k] = v }
        else
          @attrs[:cache_version][nil] = value
        end
      end

      def app(subname, &blk)
        options        = Marshal.load(Marshal.dump(attrs)) # FIXME remove, override clone method
        options[:name] = [name, subname].compact.join('-')

        self.class.new(**options).tap do |app|
          app.instance_eval(&blk) if block_given?
          @apps += app.apps
        end
      end

      def option(name, *args)
        if args.size == 0
          attrs[name]
        elsif args.size == 1
          attrs[name] = args.first
        else
          raise ArgumentError
        end
      end
    end
  end
end
