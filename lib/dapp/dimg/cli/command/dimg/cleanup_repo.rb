module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class CleanupRepo < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg cleanup repo [options] [DIMG ...] REPO

Options:
BANNER

        option :lock_timeout,
               long: '--lock-timeout TIMEOUT',
               description: 'Redefine resource locking timeout (in seconds)',
               proc: ->(v) { v.to_i }

        option :with_stages,
               long: '--with-stages',
               boolean: true

        option :without_kube,
               long: '--without-kube',
               boolean: true

        option :registry_username,
               long: '--registry-username USERNAME',
               default: ENV.key?("DAPP_DIMG_CLEANUP_REGISTRY_PASSWORD") ? : "dapp-cleanup-repo" : nil # FIXME: https://gitlab.com/gitlab-org/gitlab-ce/issues/41384

        option :registry_password,
               long: '--registry-password PASSWORD',
               default: ENV["DAPP_DIMG_CLEANUP_REGISTRY_PASSWORD"] # FIXME: https://gitlab.com/gitlab-org/gitlab-ce/issues/41384

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repo = self.class.required_argument(self, 'repo')
          run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments, repo: repo, verbose: true))
        end
      end
    end
  end
end
