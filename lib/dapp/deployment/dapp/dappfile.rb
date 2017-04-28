module Dapp
  module Deployment
    module Dapp
      module Dappfile
        def apps_configs
          @apps_configs ||= begin
            config._app.select do |app|
              apps_patterns.any? { |pattern| app._name.nil? || File.fnmatch(pattern, app._name) }
            end.tap do |apps|
              raise ::Dapp::Error::Dapp, code: :no_such_app, data: { apps_patterns: apps_patterns.join(', ') } if apps.empty?
            end
          end
        end

        def apps_patterns
          @apps_patterns ||= (options[:apps_patterns] || []).tap do |patterns|
            patterns << '*' unless patterns.any?
          end
        end
      end # Dappfile
    end # Dapp
  end # Dimg
end # Dapp
