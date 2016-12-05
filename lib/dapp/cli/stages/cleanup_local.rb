module Dapp
  class CLI
    class Stages
      # stages cleanup local subcommand
      class CleanupLocal < Base
        banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stages cleanup local [options] [DIMG ...] [REPO]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
        option :proper_cache_version,
               long: '--improper-cache-version',
               boolean: true

        option :proper_git_commit,
               long: '--improper-git-commit',
               boolean: true

        option :proper_repo_cache,
               long: '--improper-repo-cache',
               boolean: true

        def run(argv = ARGV)
          self.class.parse_options(self, argv)
          repository = repo
          Project.new(cli_options: config, dimgs_patterns: cli_arguments).send(run_method, repository)
        end

        def repo
          config[:proper_repo_cache] ? self.class.required_argument(self) : nil
        end

        def run_method
          :stages_cleanup_local
        end
      end
    end
  end
end
