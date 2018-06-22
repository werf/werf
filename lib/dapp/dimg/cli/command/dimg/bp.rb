module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Bp < Push
        banner <<BANNER.freeze
Usage:

  dapp dimg bp [options] [DIMG ...] REPO

    DIMG                        Dapp image to process [default: *].
    REPO                        Pushed image name.

Options:
BANNER

        # build options

        extend ::Dapp::Dimg::CLI::Options::Introspection
        extend ::Dapp::CLI::Options::Ssh

        option :tmp_dir_prefix,
              long: '--tmp-dir-prefix PREFIX',
              description: 'Tmp directory prefix (/tmp by default). Used for build process service directories.'

        option :lock_timeout,
              long: '--lock-timeout TIMEOUT',
              description: 'Redefine resource locking timeout (in seconds)',
              proc: ->(v) { v.to_i }

        option :build_context_directory,
              long: '--build-context-directory DIR_PATH',
              default: nil

        option :use_system_tar,
              long: '--use-system-tar',
              boolean: true,
              default: false

        # push options

        extend ::Dapp::CLI::Options::Tag

        option :with_stages,
               long: '--with-stages',
               boolean: true

        option :registry_username,
               long: '--registry-username USERNAME'

        option :registry_password,
               long: '--registry-password PASSWORD'
      end
    end
  end
end
