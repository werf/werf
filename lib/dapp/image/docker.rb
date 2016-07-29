module Dapp
  # Image
  module Image
    # Docker
    class Docker
      include Helper::Shellout

      attr_reader :from
      attr_reader :name

      def initialize(name:, from: nil)
        @from = from
        @name = name
      end

      def id
        @id || begin
          return cache[name][:exist] ? cache[name][:id] : nil if cache.include?(name)
          if shellout!("docker images -q --no-trunc=true #{name}").stdout.strip.empty?
            cache[name] = { exist: false }
          else
            reset_cache
            cache[name][:id]
          end
        end
      end

      def untag!
        raise Error::Build, code: :image_already_untagged, data: { name: name } unless tagged?
        shellout!("docker rmi #{name}")
        cache.delete(name)
      end

      def push!(log_verbose: false, log_time: false)
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        shellout!("docker push #{name}", log_verbose: log_verbose, log_time: log_time)
        reset_cache
      end

      def pull!(log_verbose: false, log_time: false)
        return if tagged?
        shellout!("docker pull #{name}", log_verbose: log_verbose, log_time: log_time)
        reset_cache
        @pulled = true
      end

      def tagged?
        !!id
      end

      def pulled?
        !!@pulled
      end

      def info
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        [cache[name][:created_at], cache[name][:size]]
      end

      protected

      class << self; attr_accessor :cache end

      def cache
        self.class.cache || reset_cache
      end

      def reset_cache
        (self.class.cache = {}).tap do
          shellout!("docker images --format='dapp:{{.Tag}};{{.ID}};{{.CreatedAt}};{{.Size}}' dapp").stdout.lines.each do |line|
            name, id, created_at, size = line.split(';')
            self.class.cache[name] = { id: id, created_at: created_at, size: size, exist: true }
          end
        end
      end
    end # Docker
  end # Image
end # Dapp
