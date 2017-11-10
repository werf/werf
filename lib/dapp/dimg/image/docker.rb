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
          self.class.image_inspect(self.name)["Id"]
        end

        def untag!
          raise Error::Build, code: :image_already_untagged, data: { name: name } unless tagged?
          dapp.shellout!("#{dapp.host_docker} rmi #{name}")
          self.class.reset_image_inspect(self.name)
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

          self.class.reset_image_inspect(self.name)
        end

        def tagged?
          not self.class.image_inspect(self.name).empty?
        end

        def created_at
          raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
          self.class.image_inspect(self.name)["Created"]
        end

        def size
          raise Error::Build, code: :image_not_exist, data: { name: name } unless tagged?
          Float(self.class.image_inspect(self.name)["Size"])
        end

        def config_option(option)
          raise Error::Build, code: :image_not_exist, data: { name: name } if built_id.nil?
          self.class.image_config_option(image_id: built_id, option: option)
        end

        class << self
          def image_inspect(image_id)
            image_inspects[image_id] ||= begin
              cmd = ::Dapp::Dapp.shellout("#{::Dapp::Dapp.host_docker} inspect --type=image #{image_id}")

              if cmd.exitstatus != 0
                if cmd.stderr.start_with? "Error: No such image:"
                  {}
                else
                  ::Dapp::Dapp.shellout_cmd_should_succeed! cmd
                end
              else
                Array(JSON.parse(cmd.stdout.strip)).first || {}
              end
            end
          end

          def image_config(image_id)
            image_inspect(image_id)["Config"] || {}
          end

          def image_config_option(image_id:, option:)
            image_config(image_id)[option]
          end

          def reset_image_inspect(image_id)
            image_inspects.delete(image_id)
          end

          def image_inspects
            @image_inspects ||= {}
          end
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
            image_inspects[tag] = image_inspect(id)
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
