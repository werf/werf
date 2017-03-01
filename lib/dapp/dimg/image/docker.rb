module Dapp
  module Dimg
    # Image
    module Image
      # Docker
      class Docker
        attr_reader :from
        attr_reader :name
        attr_reader :dapp

        def self.image_by_name(name:, **kwargs)
          (@images ||= {})[name] ||= new(name: name, **kwargs)
        end

        def initialize(name:, dapp:, from: nil)
          @from = from
          @name = name
          @dapp = dapp
        end

        def id
          cache[:id]
        end

        def untag!
          raise Error::Build, code: :image_already_untagged, data: { name: name } unless tagged?
          dapp.shellout!("docker rmi #{name}")
          cache_reset
        end

        def push!
          raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
          dapp.log_secondary_process(dapp.t(code: 'process.image_push', data: { name: name })) do
            dapp.shellout!("docker push #{name}", log_verbose: true)
          end
        end

        def pull!
          return if tagged?
          dapp.log_secondary_process(dapp.t(code: 'process.image_pull', data: { name: name })) do
            dapp.shellout!("docker pull #{name}", log_verbose: true)
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
          output = ::Dapp::Dapp.shellout!("docker inspect --format='{{json .Config.#{option.to_s.capitalize}}}' #{image_id}").stdout.strip
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
            ::Dapp::Dapp.shellout!("docker tag #{id} #{tag}")
            cache_reset
          end

          def cache
            @cache ||= (@cache = {}).tap { cache_reset }
          end

          def cache_reset(name = '')
            cache.delete(name)
            ::Dapp::Dapp.shellout!("docker images --format='{{.Repository}}:{{.Tag}};{{.ID}};{{.CreatedAt}};{{.Size}}' --no-trunc #{name}")
              .stdout
              .lines
              .each do |l|
              name, id, created_at, size_field = l.split(';')
              size = begin
                number, unit = size_field.split
                coef = case unit.to_s.downcase
                       when 'b'  then return number.to_f
                       when 'kb' then 1
                       when 'mb' then 2
                       when 'gb' then 3
                       when 'tb' then 4
                       end
                number.to_f * (1000**coef)
              end
              cache[name] = { id: id, created_at: created_at, size: size }
            end
          end
        end
      end # Docker
    end # Image
  end # Dimg
end # Dapp
