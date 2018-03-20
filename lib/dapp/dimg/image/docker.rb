module Dapp
  module Dimg
    module Image
      class Docker
        attr_reader :from
        attr_reader :name
        attr_reader :dapp

        class << self
          def image_by_name(name:, **kwargs)
            images[name] ||= new(name: name, **kwargs)
          end

          def image_reset(name)
            images.delete(name)
          end

          def images
            @images ||= {}
          end
        end

        def initialize(name:, dapp:, from: nil)
          @from = from
          @name = name
          @dapp = dapp
        end

        def id
          image_inspect["Id"]
        end

        def reset_image_inspect
          @image_inspect = nil
        end

        def image_inspect
          @image_inspect ||= begin
            res = self.dapp.ruby2go_image(image: JSON.dump(name: name), command: "inspect")
            raise res["error"] if res["error"]
            image = JSON.load(res['data']['image'])
            image['image_inspect'] || {}
          end
        end

        def untag!
          raise Error::Build, code: :image_already_untagged, data: { name: name } unless tagged?
          dapp.shellout!("#{dapp.host_docker} rmi #{name}")
          reset_image_inspect
        end

        def push!
          raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
          dapp.log_secondary_process(dapp.t(code: 'process.image_push', data: { name: name })) do
            dapp.shellout!("#{dapp.host_docker} push #{name}", verbose: true)
          end
        end

        def pull!
          return if tagged?
          dapp.log_secondary_process(dapp.t(code: 'process.image_pull', data: { name: name })) do
            dapp.shellout!("#{dapp.host_docker} pull #{name}", verbose: true)
          end
          reset_image_inspect
        end

        def tagged?
          not image_inspect.empty?
        end

        protected

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
            ::Dapp::Dapp.shellout!("#{::Dapp::Dapp.host_docker} tag #{id} #{tag}", verbose: verbose, quiet: quiet)
          end

          def save!(image_or_images, file_path, verbose: false, quiet: false)
            images = Array(image_or_images).join(' ')
            ::Dapp::Dapp.shellout!("#{::Dapp::Dapp.host_docker} save -o #{file_path} #{images}", verbose: verbose, quiet: quiet)
          end

          def load!(file_path, verbose: false, quiet: false)
            ::Dapp::Dapp.shellout!("#{::Dapp::Dapp.host_docker} load -i #{file_path}", verbose: verbose, quiet: quiet)
          end
        end
      end # Docker
    end # Image
  end # Dimg
end # Dapp
