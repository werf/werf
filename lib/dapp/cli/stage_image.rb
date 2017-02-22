module Dapp
  class CLI
    # CLI stage image subcommand
    class StageImage < Base
      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Usage:
  dapp stage image [options] [DIMG]

    DIMG                        Dapp image to process [default: *].

Options:
BANNER
      option :stage,
             long: '--stage STAGE',
             proc: proc { |v| v.to_sym },
             default: :docker_instructions,
             in: [:from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install,
                  :g_a_post_install_patch, :after_install_artifact, :before_setup, :before_setup_artifact, :g_a_pre_setup_patch,
                  :setup, :g_a_post_setup_patch, :after_setup_artifact, :g_a_latest_patch, :docker_instructions]
    end
  end
end
