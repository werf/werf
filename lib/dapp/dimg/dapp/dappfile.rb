module Dapp
  module Dimg
    module Dapp
      module Dappfile
        def nameless_dimg?
          dimgs_names.first.nil?
        end

        def dimg_name!
          one_dimg!
          build_configs.first._name
        end

        def one_dimg!
          return if build_configs.one?
          raise ::Dapp::Error::Command, code: :command_unexpected_dimgs_number, data: { dimgs_names: build_configs.map(&:_name).join('`, `') }
        end

        def dimgs_names
          build_configs.map(&:_name)
        end

        def build_configs
          @build_configs ||= begin
            config._dimg.select do |dimg|
              dimgs_patterns.any? { |pattern| dimg._name.nil? || File.fnmatch(pattern, dimg._name) }
            end.tap do |dimgs|
              raise ::Dapp::Error::Dapp, code: :no_such_dimg, data: { dimgs_patterns: dimgs_patterns.join('`, `') } if dimgs.empty?
            end
          end
        end

        def dimgs_patterns
          @dimgs_patterns ||= begin
            (options[:dimgs_patterns] || []).tap do |dimgs_patterns|
              dimgs_patterns << '*' unless dimgs_patterns.any?
            end
          end
        end
      end # Dappfile
    end # Dapp
  end # Dimg
end # Dapp