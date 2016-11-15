module Dapp
  # Image
  module Image
    # Docker
    class Docker
      attr_reader :from
      attr_reader :name
      attr_reader :project

      def initialize(name:, project:, from: nil)
        @from = from
        @name = name
        @project = project
      end

      def id
        cache[:id]
      end

      def untag!
        raise Error::Build, code: :image_already_untagged, data: { name: name } unless tagged?
        project.shellout!("docker rmi #{name}")
        cache_reset
      end

      def push!
        raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
        project.log_secondary_process(project.t(code: 'process.image_push', data: { name: name })) do
          project.shellout!("docker push #{name}", log_verbose: true)
        end
      end

      def pull!
        return if tagged?
        project.log_secondary_process(project.t(code: 'process.image_pull', data: { name: name })) do
          project.shellout!("docker pull #{name}", log_verbose: true)
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

      def self.image_config_option(image_id:, option:)
        output = Project.shellout!("docker inspect --format='{{json .Config.#{option.to_s.capitalize}}}' #{image_id}").stdout.strip
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
        def image_regex
          /^[a-z0-9]+(?:[._-][a-z0-9]+)*(:[\w][\w.-]{0,127})?$/
        end

        def tag!(id:, tag:)
          Project.shellout!("docker tag #{id} #{tag}")
          cache_reset
        end

        def cache
          @cache ||= (@cache = {}).tap { cache_reset }
        end

        def cache_reset(name = '')
          cache.delete(name)
          Project.shellout!("docker images --format='{{.Repository}}:{{.Tag}};{{.ID}};{{.CreatedAt}};{{.Size}}' --no-trunc #{name}").stdout.lines.each do |l|
            name, id, created_at, size = l.split(';')
            cache[name] = { id: id, created_at: created_at, size: size }
          end
        end
      end
    end # Docker
  end # Image
end # Dapp
