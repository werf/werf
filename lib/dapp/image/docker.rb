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
        cache[:id]
      end

      def untag!
        raise Error::Build, code: :image_already_untagged, data: { name: name } unless tagged?
        shellout!("docker rmi #{name}")
        cache_reset
      end

      def push!(log_verbose: false, log_time: false)
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        shellout!("docker push #{name}", log_verbose: log_verbose, log_time: log_time)
      end

      def pull!(log_verbose: false, log_time: false)
        return if tagged?
        project.with_log_indent do
          project.log_secondary_process(project.t(code: 'process.image_pull', data: { name: name }), short: !log_verbose) do
            shellout!("docker pull #{name}", log_verbose: log_verbose, log_time: log_time)
          end
        end
        cache_reset
      end

      def tagged?
        !!id
      end

      def created_at
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        cache[:created_at]
      end

      def size
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        cache[:size]
      end

      def labels
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        self.class.image_config_option(image_id: id, option: 'labels')
      end

      def self.image_config_option(image_id:, option:)
        output = shellout!("docker inspect --format='{{json .Config.#{option.to_s.capitalize}}}' #{image_id}").stdout.strip
        output == 'null' ? [] : JSON.parse(output)
      end

      def cache_reset
        self.class.cache_reset(name)
      end

      protected

      def cache
        self.class.cache[name.to_s] || {}
      end

      class << self
        def cache
          @cache ||= (@cache = {}).tap { cache_reset }
        end

        def cache_reset(name = '')
          cache.delete(name)
          shellout!("docker images --format='{{.Repository}}:{{.Tag}};{{.ID}};{{.CreatedAt}};{{.Size}}' --no-trunc #{name}").stdout.lines.each do |line|
            name, id, created_at, size = line.split(';')
            cache[name] = { id: id, created_at: created_at, size: size }
          end
        end
      end
    end # Docker
  end # Image
end # Dapp
