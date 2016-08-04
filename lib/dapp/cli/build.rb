require 'mixlib/cli'

module Dapp
  class CLI
    # CLI build subcommand
    class Build < Base
      include Dapp::Helper::Shellout

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp build [options] [PATTERN ...]

    PATTERN                     Applications to process [default: *].

Options:
BANNER
      option :tmp_dir_prefix,
             long: '--tmp-dir-prefix PREFIX',
             description: 'Tmp directory prefix'

      option :metadata_dir,
             long: '--metadata-dir PATH',
             description: 'Metadata directory'

      option :git_artifact_branch,
             long: '--git-artifact-branch BRANCH',
             description: 'Default branch to archive artifacts from'

      option :introspect_error,
             long: '--introspect-error',
             boolean: true,
             default: false

      option :introspect_before_error,
             long: '--introspect-before-error',
             boolean: true,
             default: false

      option :introspect_stage,
             long: '--introspect-stage STAGE',
             proc: proc { |v| v.to_sym },
             in: [nil, :from, :infra_install, :source_1_archive, :source_1, :app_install, :source_2,
                  :infra_setup, :source_3, :chef_cookbooks, :app_setup, :source_4, :source_5]

      def run(*args)
        super
      rescue Exception::IntrospectImage => e
        $stderr.puts(e.net_status[:message])
        data = e.net_status[:data]
        system("docker run -ti --rm #{data[:options]} #{data[:built_id]} bash").tap do |res|
          shellout("docker rmi #{data[:built_id]}") if data[:rmi]
          res || raise(Dapp::Error::Application, code: :application_not_run)
        end
      end
    end
  end
end
