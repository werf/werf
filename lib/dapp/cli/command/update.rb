module Dapp
  class CLI
    module Command
      class Update < ::Dapp::CLI
        def run(_argv)
          spec = Gem::Specification.find do |s|
            File.fnmatch(File.join(s.full_gem_path, '*'), __FILE__)
          end
          Gem.install(spec.name, spec.version.approximate_recommendation)
        end
      end
    end
  end
end
