module Dapp
  module Config
    class Main < Base
      def initialize(**options)
        keys.merge!(options)

        if options[:name]
          keys[:basename] = [keys[:basename], options[:name]].compact.join('-')
          keys[:name] = nil
        elsif options[:dappfile_path]
          keys[:basename] ||= Pathname.new(options[:dappfile_path]).expand_path.parent.basename
        end

        keys[:home_path] ||= Pathname.new(options[:dappfile_path] || 'fakedir').parent.expand_path.to_s

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
        raise unless keys[another_builder].nil? && keys[:builder] == builder_name
      end

      def name
        keys[:basename]
      end

      def apps
        @apps.empty? ? [self] : @apps.flatten
      end

      private

      def keys
        @keys ||= {}
      end

      def app(name, &blk)
        options = Marshal.load(Marshal.dump(keys))
        options[:name] = name

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
