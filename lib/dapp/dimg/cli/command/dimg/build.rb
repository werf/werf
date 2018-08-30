module Dapp::Dimg::CLI
  module Command
    class Dimg < ::Dapp::CLI
      class Build < Base
        banner <<BANNER.freeze
Usage:

  dapp dimg build [options] [DIMG ...]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER

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

        option :force_save_cache,
               long: '--force-save-cache',
               boolean: true,
               default: false
      end
    end
  end
end
