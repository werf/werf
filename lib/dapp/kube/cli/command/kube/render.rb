module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class Render < Base
      banner <<BANNER.freeze
Usage:

  dapp kube render [options] [REPO]

Options:
BANNER
      extend ::Dapp::CLI::Options::Tag

      option :namespace,
             long: '--namespace NAME',
             default: nil

      option :context,
             long: '--context NAME',
             default: nil

      option :tmp_dir_prefix,
             long: '--tmp-dir-prefix PREFIX',
             description: 'Tmp directory prefix (/tmp by default). Used for build process service directories.'

      option :helm_set_options,
             long: '--set STRING_ARRAY',
             default: [],
             proc: proc { |v| composite_options(:helm_set) << v }

      option :helm_values_options,
             long: '--values FILE_PATH',
             default: [],
             proc: proc { |v| composite_options(:helm_values) << v }

      option :helm_secret_values_options,
             long: '--secret-values FILE_PATH',
             default: [],
             proc: proc { |v| composite_options(:helm_secret_values) << v }

      option :templates,
             long: '--template GLOB_PATTERN',
             short: '-t GLOB_PATTERN',
             default: [],
             proc: proc { |v| composite_options(:templates) << v }

      option :registry_username,
             long: '--registry-username USERNAME'

      option :registry_password,
             long: '--registry-password PASSWORD'

      option :without_registry,
              long: "--without-registry",
              default: false,
              boolean: true,
              description: "Do not connect to docker registry to obtain docker image ids of dimgs being deployed."

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(nil, options: cli_options) do |dapp|
          repo = if not cli_arguments[0].nil?
            self.class.required_argument(self, 'repo')
          else
            dapp.name
          end

          dapp.options[:repo] = repo

          dapp.public_send(run_method)
        end
      end

      def log_running_time
        false
      end
    end
  end
end
