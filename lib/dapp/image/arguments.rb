module Dapp
  module Image
    # Arguments
    module Arguments
      def add_change_volume(value)
        add_change_option(:volume, value)
      end

      def add_change_expose(value)
        add_change_option(:expose, value)
      end

      def add_change_env(**options)
        add_change_option(:env, options_to_args(options))
      end

      def add_change_label(**options)
        add_change_option(:label, options_to_args(options))
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
        add_change_option(:workdir, value)
      end

      def add_change_user(value)
        add_change_option(:user, value)
      end

      def add_cmd(value)
        add_option(:cmd, value)
      end

      def add_volume(value)
        add_option(:volume, value)
      end

      def add_volumes_from(value)
        add_option(:'volumes-from', value)
      end

      def add_entrypoint(value)
        add_option(:entrypoint, value)
      end

      def add_commands(*commands)
        @bash_commands.concat(commands.flatten)
      end

      protected

      attr_reader :bash_commands
      attr_reader :options, :change_options

      def add_option(key, value)
        add_option_default(options, key, value)
      end

      def add_change_option(key, value)
        add_option_default(change_options, key, value)
      end

      def add_option_default(hash, key, value)
        hash[key] = (hash[key].nil? ? value : (Array(hash[key]) << value).flatten)
      end

      def from_options
        return {} if from.nil?
        [:entrypoint, :cmd].each_with_object({}) do |option, options|
          output = shellout!("docker inspect --format='{{json .Config.#{option.to_s.capitalize}}}' #{from.built_id}").stdout.strip
          options[option] = output == 'null' ? [] : JSON.parse(output)
          options
        end
      end

      def options_to_args(options)
        options.map { |key, value| "#{key}=#{value}" }
      end

      def prepared_options
        prepared_options_default(options) { |k, vals| Array(vals).map { |v| "--#{k}=#{v}" }.join(' ') }
      end

      def prepared_change
        prepared_options_default(from_options.merge(change_options)) do |k, vals|
          if [:cmd, :entrypoint].include? k
            %(-c '#{k.to_s.upcase} #{Array(vals)}')
          else
            Array(vals).map { |v| %(-c "#{k.to_s.upcase} #{v}") }.join(' ')
          end
        end
      end

      def prepared_options_default(hash)
        hash.map { |k, vals| yield(k, vals) }.join(' ')
      end

      def prepared_bash_command
        shellout_pack prepared_commands.join(' && ')
      end

      def prepared_commands
        return ['true'] if bash_commands.empty?
        bash_commands
      end
    end
  end # Image
end # Dapp
