module Dapp
  class CLI
    module Command
      class Update < ::Dapp::CLI
        def run(_argv)
          spec = Gem::Specification.find do |s|
            File.fnmatch(File.join(s.full_gem_path, '*'), __FILE__)
          end
          unless (latest_version = latest_beta_version(spec)).nil?
            Gem.install(spec.name, latest_version)
          end
        rescue Gem::FilePermissionError => e
          raise Errno::EACCES, e.message
        end

        def latest_beta_version(current_spec)
          minor_version = current_spec.version.approximate_recommendation
          url = "https://rubygems.org/api/v1/versions/#{current_spec.name}.json"
          response = Excon.get(url)
          if response.status == 200
            JSON.parse(response.body)
              .reduce([]) { |versions, spec| versions << Gem::Version.new(spec['number']) }
              .reject { |spec_version| minor_version != spec_version.approximate_recommendation || current_spec.version >= spec_version }
              .first
          else
            raise ::Dapp::Error::Base, message: "Cannot get `#{url}`: got bad http status `#{response.status}`!"
          end
        end
      end
    end
  end
end
