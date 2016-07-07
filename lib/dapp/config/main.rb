module Dapp
  module Config
    class Main < Base
      def initialize(**options)
        @keys = options

        @keys[:home_path]     ||= Pathname.new(@keys[:dappfile_path]).parent.expand_path.to_s
        @keys[:name]          ||= Pathname.new(@keys[:home_path]).basename unless @keys[:name]

        @keys[:builder]       ||= :shell
        @keys[:cache_version] ||= {}
        @apps                 = []

        super()
      end

      def method_missing(name, *args)
        return keys[name] if keys.key?(name)
        klass      = Config.const_get(name.to_s.split('_').map(&:capitalize).join)
        keys[name] ||= klass.new(self, *args)
      rescue NameError
        super
      end

      def builder_validation(builder_name)
        raise RuntimeError, "Builder type '#{builder_name}' is not defined!" unless keys[:builder] == builder_name
      end

      def name(*args)
        option(:name, *args)
      end

      def builder(*args)
        option(:builder, *args)
      end

      def apps
        @apps.empty? ? [self] : @apps.flatten
      end

      def cache_key(key = nil)
        @keys[:cache_version][key]
      end

      private

      def keys
        @keys ||= {}
      end

      def cache_version(value = nil, **options)
        if value.nil?
          options.each { |k, v| @keys[:cache_version][k] = v }
        else
          @keys[:cache_version][nil] = value
        end
      end

      def app(subname, &blk)
        options        = Marshal.load(Marshal.dump(keys))
        options[:name] = [name, subname].compact.join('-')

        self.class.new(**options).tap do |app|
          app.instance_eval(&blk) if block_given?
          @apps += app.apps
        end
      end

      def option(name, *args)
        if args.size == 0
          keys[name]
        elsif args.size == 1
          keys[name] = args.first
        else
          raise ArgumentError
        end
      end
    end
  end
end
