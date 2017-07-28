module Dapp
  class CLI
    module Command
      class Update < ::Dapp::CLI
        def run(_argv)
          spec = Gem::Specification.find do |s|
            File.fnmatch(File.join(s.full_gem_path, '*'), __FILE__)
          end
          Gem.install(spec.name, approximate_recommendation(spec.version))
        rescue Gem::FilePermissionError => e
          raise Errno::EACCES, e.message
        end

        # get latest beta-version
        def approximate_recommendation(version)
          [version.approximate_recommendation, 0].join('.')
        end
      end
    end
  end
end
