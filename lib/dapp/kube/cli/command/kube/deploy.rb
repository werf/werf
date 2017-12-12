module Dapp::Kube::CLI::Command
  class Kube < ::Dapp::CLI
    class Deploy < Base
      banner <<BANNER.freeze
Usage:

  dapp deployment deploy [options] [REPO]

Options:
BANNER
      option :namespace,
             long: '--namespace NAME',
             default: nil

      option :image_version,
             long: '--image-version TAG',
             description: "Custom tag (alias for --tag)",
             default: [],
             proc: proc { |v| composite_options(:image_versions) << v }

      option :tag,
             long: '--tag TAG',
             description: 'Custom tag',
             default: [],
             proc: proc { |v| composite_options(:tags) << v }

      option :tag_branch,
             long: '--tag-branch',
             description: 'Tag by git branch',
             boolean: true

      option :tag_build_id,
             long: '--tag-build-id',
             description: 'Tag by CI build id',
             boolean: true

      option :tag_ci,
             long: '--tag-ci',
             description: 'Tag by CI branch and tag',
             boolean: true

      option :tag_commit,
             long: '--tag-commit',
             description: 'Tag by git commit',
             boolean: true

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

      option :timeout,
             long: '--timeout INTEGER_SECONDS',
             default: nil,
             description: 'Default timeout to wait for resources to become ready, 300 seconds by default.',
             proc: proc {|v| Integer(v)}

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
        run_dapp_command(run_method, options: cli_options)
      end

      def before_dapp_run_command(dapp, &blk)
        super(dapp) do
          yield if block_given?

          # Опция repo определяется в данном хуке, чтобы установить
          # значение по умолчанию из объекта dapp: dapp.name
          repo = if not cli_arguments[0].nil?
            self.class.required_argument(self, 'repo')
          else
            dapp.name
          end
          dapp.options[:repo] = repo

          dapp.options[:tag] = [*dapp.options.delete(:tag), *dapp.options.delete(:image_version)]
        end # super
      end

    end
  end
end
