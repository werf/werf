module Dapp
  module Dimg
    module Image
      module Argument
        def add_change_volume(value)
          add_change_option(:volume, value)
        end

        def add_change_expose(value)
          add_change_option(:expose, value)
        end

        def add_change_env(**options)
          (change_options[:env] ||= {}).merge!(options)
        end

        def add_change_label(**options)
          (change_options[:label] ||= {}).merge!(options)
        end

        def add_change_cmd(value)
          add_change_option(:cmd, value)
        end

        def add_change_entrypoint(value)
          add_change_option(:entrypoint, value)
        end

        def add_change_onbuild(value)
          add_change_option(:onbuild, value)
        end

        def add_change_workdir(value)
          change_options[:workdir] = value
        end

        def add_change_user(value)
          change_options[:user] = value
        end

        def add_service_change_label(**options)
          (service_change_options[:label] ||= {}).merge!(options)
        end

        def add_env(**options)
          (self.options[:env] ||= {}).merge!(options)
        end

        def add_volume(value)
          add_option(:volume, value)
        end

        def add_volumes_from(value)
          add_option(:'volumes-from', value)
        end

        def add_command(*commands)
          @bash_commands.concat(commands.flatten)
        end

        def add_service_command(*commands)
          @service_bash_commands.concat(commands.flatten)
        end

        def prepare_instructions(options)
          options.map do |key, vals|
            case key
            when :workdir, :user
              [vals]
            when :cmd, :entrypoint
              vals = [''] if vals == [] && ::Dapp::Dapp.host_docker_minor_version >= Gem::Version.new('17.10')
              [vals]
            when :env, :label then options_to_args(vals)
            else vals
            end.map { |val| %(#{key.to_s.upcase} #{val}) }
          end.flatten
        end

        protected

        attr_reader :bash_commands, :service_bash_commands
        attr_reader :change_options
        attr_reader :service_change_options
        attr_reader :options

        def add_option(key, value)
          add_option_default(options, key, value)
        end

        def add_change_option(key, value)
          add_option_default(change_options, key, value)
        end

        def add_service_change_option(key, value)
          add_option_default(service_change_options, key, value)
        end

        def add_option_default(hash, key, value)
          hash[key] = (hash[key].nil? ? [value] : (hash[key] << value)).flatten
        end

        def options_to_args(options)
          options.map { |key, value| "#{key}=#{value}" }
        end
      end
    end # Image
  end # Dimg
end # Dapp
