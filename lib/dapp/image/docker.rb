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
        def image_name_format
          separator = '[_.]|__|[-]*'
          tag = "[[:alnum:]][[[:alnum:]]#{separator}]{0,127}"
          "#{DockerRegistry.repo_name_format}(:(?<tag>#{tag}))?"
        end

        def image_name?(name)
          !(/^#{image_name_format}$/ =~ name).nil?
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
            name, id, created_at, size_field = l.split(';').map(&:strip)
            size = begin
              match = size_field.match(/^(\d+(\.\d+)?)\ ?(b|kb|mb|gb|tb)$/i)
              raise Error::Build, code: :unsupported_docker_image_size_format, data: {value: size_field} unless match and match[1] and match[3]

              number = match[1].to_f
              unit = match[3].downcase

              coef = case unit
                     when 'b'  then 0
                     when 'kb' then 1
                     when 'mb' then 2
                     when 'gb' then 3
                     when 'tb' then 4
                     end

              number * (1000**coef)
            end
            cache[name] = { id: id, created_at: created_at, size: size }
          end
        end
      end
    end # Docker
  end # Image
end # Dapp
