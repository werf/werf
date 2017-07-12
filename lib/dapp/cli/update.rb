module Dapp
  class CLI
    class Update < ::Dapp::CLI
      def run(_argv)
        spec = Gem::Specification.find { |s| File.fnmatch(File.join(s.full_gem_path, '*'), __FILE__) }
        Gem.install(spec.name, approximate_recommendation(spec.version))
      end

      # get latest beta-version
      def approximate_recommendation(version)
        [version.approximate_recommendation, 0].join('.')
      end
    end
  end
end
