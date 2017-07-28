module Dapp
  class Dapp
    module DappConfig
      SUPPORTED_CONFIG_OPTIONS = {
        verbose:   [FalseClass, TrueClass],
        quiet:     [FalseClass, TrueClass],
        dev:       [FalseClass, TrueClass],
        time:      [FalseClass, TrueClass],
        dry_run:   [FalseClass, TrueClass],
        build_dir: [String],
        color:     [String]
      }

      SUPPORTED_CONFIG_OPTIONS.keys.each do |opt|
        define_method "option_#{opt}" do
          options[opt] || config_options[opt]
        end
      end

      def option_color
        if options[:color] == 'auto'
          config_options[:color] || 'auto'
        else
          options[:color]
        end
      end

      def option_dev
        if options[:dev].nil?
          config._dev_mode || config_options[:dev]
        else
          options[:dev]
        end
      end

      def config_options
        @config_options ||= begin
          config_search_paths = []
          config_search_paths << File.join(Dir.home)
          config_search_paths << path if dappfile_exists?

          config_search_paths.reduce({}) do |options, dir|
            if (config_options_path = make_path(dir, '.dapp_config')).file?
              config_options = begin
                yaml_load_file(config_options_path).tap do |c_options|
                  c_options.merge!(c_options.in_depth_merge(c_options['ci'] || {})) if ENV['GITLAB_CI'] || ENV['TRAVIS']
                  c_options.delete('ci')
                end
              rescue Psych::SyntaxError => e
                raise Error::Dapp, code: :dapp_config_file_incorrect, data: { message: e.message }
              end
              options.in_depth_merge(config_options)
            else
              options
            end
          end.symbolize_keys
        end
      end

      def validate_config_options!
        data_list_format = proc { |list| list.map { |e| "'#{e}'" }.join(', ') }

        unless (unsupported_keys = config_options.select { |k, _| !SUPPORTED_CONFIG_OPTIONS.keys.include?(k) }.keys).empty?
          log_warning(desc: { code: :unsupported_dapp_config_options,
                              data: { options: data_list_format.call(unsupported_keys),
                                      supported_options: data_list_format.call(SUPPORTED_CONFIG_OPTIONS.keys) } })
        end

        config_options.each do |k, v|
          next unless SUPPORTED_CONFIG_OPTIONS.keys.include?(k)

          if k == :color
            raise Error::Dapp,
                  code: :incorrect_dapp_config_option_color,
                  data: { value: v, expected: data_list_format.call(%w(auto on off)) } unless %w(auto on off).member?(v)
          elsif !SUPPORTED_CONFIG_OPTIONS[k].member? v.class
            raise Error::Dapp,
                  code: :incorrect_dapp_config_option,
                  data: { option: k, value: v, expected: data_list_format.call(SUPPORTED_CONFIG_OPTIONS[k]) }
          end
        end
      end
    end # DappConfig
  end # Dapp
end # Dapp
