module Dapp
  module Dimg
    module Image
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
          dapp.docker_client.image_remove(name)
          cache_reset
        end

        def push!
          raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
          dapp.log_secondary_process(dapp.t(code: 'process.image_push', data: { name: name })) do
            dapp.docker_client.image_push(name)
          end
        end

        def pull!
          return if tagged?
          dapp.docker_client.image_pull(name)
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
          ::Dapp::Dapp.docker_client.image(image_id).json['Config'][option.to_s.capitalize] || []
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
            "#{DockerRegistry.repo_name_format}(:(?<tag>#{tag_format}))?"
          end

          def tag_format
            '(?![-.])[a-zA-Z0-9_.-]{1,127}'
          end

          def image_name?(name)
            !(/^#{image_name_format}$/ =~ name).nil?
          end

          def tag?(name)
            !(/^#{tag_format}$/ =~ name).nil?
          end

          def tag!(id:, tag:, verbose: false, quiet: false)
            repo, tag = tag.split(':')
            ::Dapp::Dapp.docker_client.image_tag(verbose: false, quiet: false, name: id, repo: repo, tag: tag)
            cache_reset
          end

          def cache
            @cache ||= (@cache = {}).tap { cache_reset }
          end

          def cache_reset(name = '')
            cache.delete(name)
            ::Dapp::Dapp.docker_client.images.each do |image|
              image_info = image.info
              Array(image_info['RepoTags']).each do |repo_tag|
                cache[repo_tag] = { id: image_info['id'], created_at: image_info['Created'], size: image_info['Size'] }
              end
            end
          end
        end
      end # Docker
    end # Image
  end # Dimg
end # Dapp
