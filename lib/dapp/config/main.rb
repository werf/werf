module Dapp
  module Config
    class Main < Base
      def initialize(**options)
        keys = options

        # FIXME we always have dappfile_path
        keys[:home_path] ||= Pathname.new(keys[:dappfile_path] || 'fakedir').parent.expand_path.to_s

        unless keys[:name]
          keys[:name] ||= Pathname.new(keys[:home_path]).basename
        end

        @apps ||= []
        keys[:builder] = :shell

        super()
      end

      def method_missing(name, *args)
        return keys[name] if keys.key?(name)
        klass = Config.const_get(name.to_s.split('_').map(&:capitalize).join)
        keys[name] ||= klass.new(self, *args)
      rescue NameError
        super
      end

      def builder_validation(builder_name)
        another_builder = [:chef, :shell].find { |n| n != builder_name }
        # FIXME user friendly raise
        raise unless keys[another_builder].nil? && keys[:builder] == builder_name
      end

      def name
      end

      def name(*args)
        if args.size == 0
          keys[:name]
        elsif args.size == 1
          keys[:name] = args.first
        else
          # FIXME user friendly raise
          raise
        end
      end

      def apps
        @apps.empty? ? [self] : @apps.flatten
      end

      private

      def keys
        @keys ||= {}
      end

      def app(subname, &blk)
        options = Marshal.load(Marshal.dump(keys))
        options[:name] = [name, subname].compact.join('-')

        self.class.new(**options).tap do |app|
          app.instance_eval(&blk)
          @apps += app.apps
        end
      end

      def builder(name)
        keys[:builder] = name
      end
    end
  end
end
