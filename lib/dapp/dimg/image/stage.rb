module Dapp
  module Dimg
    module Image
      class Stage
        include Argument

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

          def save!(dapp, image_or_images, file_path, verbose: false, quiet: false)
            ruby2go_command(dapp, command: :save, options: { images: Array(image_or_images), file_path: file_path })
          end

          def load!(dapp, file_path, verbose: false, quiet: false)
            ruby2go_command(dapp, command: :load, options: { file_path: file_path })
          end

          def ruby2go_command(dapp, command:, **options)
            (options[:options] ||= {}).merge!(host_docker_config_dir: dapp.class.host_docker_config_dir)
            dapp.ruby2go_image({ command: command }.merge(options)).tap do |res|
              raise Error::Build, code: :ruby2go_image_command_failed_unexpected_error, data: { command: command, message: res["error"] } unless res["error"].nil?
              break res['data']
            end
          end
        end

        def initialize(name:, dapp:, built_id: nil, from: nil)
          @built_id = built_id

          @bash_commands          = []
          @service_bash_commands  = []
          @options                = {}
          @change_options         = {}
          @service_change_options = {}

          @from = from
          @name = name
          @dapp = dapp
        end

        def tagged?
          not image_inspect.empty?
        end

        def built?
          !built_id.nil?
        end

        def built_id
          @built_id ||= id
        end

        def id
          image_inspect["Id"]
        end

        def labels
          built_image_inspect!.fetch('Config', {}).fetch('Labels', {}) || {}
        end

        def created_at
          built_image_inspect!["Created"]
        end

        def size
          Float(built_image_inspect!["Size"])
        end

        def built_image_inspect!
          built_image_inspect
        end

        def built_image_inspect
          @built_image_inspect || image_inspect
        end

        def reset_image_inspect
          @image_inspect = nil
        end

        def image_inspect
          ruby2go_command(:inspect, extended_image_option: false) if @image_inspect.nil?
          @image_inspect
        end

        def pull!
          dapp.log_secondary_process(dapp.t(code: 'process.image_pull', data: { name: name })) do
            ruby2go_command(:pull)
          end
        end

        def push!
          dapp.log_secondary_process(dapp.t(code: 'process.image_push', data: { name: name })) do
            ruby2go_command(:push)
          end
        end

        def build!
          res = self.dapp.ruby2go_image(**ruby2go_image_build_options)
          if res["error"].nil?
            set_ruby2go_state_hash(JSON.load(res['data']['image']))
          elsif res["error"].start_with? "container run failed"
            raise Error::Build, code: :ruby2go_image_command_failed, data: { command: "build" }
          else
            raise Error::Build, code: :ruby2go_image_command_failed_unexpected_error, data: { command: "build", message: res["error"] }
          end
        end

        def ruby2go_image_build_options
          {
            image: ruby2go_image_option,
            command: :build,
            options: {
              introspection: {
                before: dapp.introspect_before_error?,
                after: dapp.introspect_error?
              },
              host_docker_config_dir: dapp.class.host_docker_config_dir,
            }
          }
        end

        def introspect!
          ruby2go_command(:introspect)
        end

        def export!(name)
          ruby2go_command(:export, options: { name: name })
        end

        def tag!(name)
          ruby2go_command(:tag, options: { name: name })
        end

        def import!(name)
          ruby2go_command(:import, options: { name: name })
        end

        def save_in_cache!
          ruby2go_command(:save_in_cache)
        end

        def untag!
          ruby2go_command(:untag)
        end

        def ruby2go_command(command, extended_image_option: true, options: {})
          command_options = ruby2go_command_options(command, extended_image_option: extended_image_option).in_depth_merge(options: options)
          self.class.ruby2go_command(dapp, **command_options).tap do |data|
            set_ruby2go_state_hash(JSON.load(data['image']))
          end
        end

        def ruby2go_command_options(command, extended_image_option: true)
          image = begin
            if extended_image_option
              ruby2go_image_option
            else
              JSON.dump({name: name})
            end
          end

          {
            image: image,
            command: command,
          }
        end

        def ruby2go_image_option
          JSON.dump(get_ruby2go_state_hash)
        end

        def get_ruby2go_state_hash
          [
            :name,
            :from,
            :built_id,
            :built_image_inspect,
            :image_inspect,
            :bash_commands,
            :service_bash_commands,
            :options,
            :change_options,
            :service_change_options,
          ].map do |name|
            if name == :from
              [name, from.get_ruby2go_state_hash] unless from.nil?
            elsif name == :built_image_inspect && built_image_inspect.empty?
            elsif name == :image_inspect && image_inspect.empty?
            else
              [name, send(name)]
            end
          end
            .compact
            .to_h
        end

        def set_ruby2go_state_hash(state_hash)
          state_hash.each do |name, value|
            variable = "@#{name}".to_sym

            case name
            when "from"
              from.set_ruby2go_state_hash(value) unless from.nil? || value.nil?
            when "built_id"
              if value.to_s.empty?
                @built_id = nil
              else
                @built_id = value
              end
            when "image_inspect"
              instance_variable_set(variable, (value || {}))
            when "options", "change_options", "service_change_options"
              instance_variable_set(variable, (value || {}).reject { |_, v| v.nil? || v.empty? }.symbolize_keys)
            when "bash_commands", "service_bash_commands"
              instance_variable_set(variable, value || [])
            else
              instance_variable_set(variable, value)
            end
          end
        end
      end # Stage
    end # Image
  end # Dimg
end # Dapp
