module Dapp::Dimg::CLI
  module Command
    class Base < ::Dapp::CLI::Command::Base
      DIMG_STAGES = [
        :from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install,
        :g_a_post_install_patch, :after_install_artifact, :before_setup, :before_setup_artifact,
        :g_a_pre_setup_patch, :setup, :g_a_post_setup_patch, :after_setup_artifact, :g_a_latest_patch, :docker_instructions
      ].freeze

      STAGE_PROC = proc do |stages|
        proc { |val| val.to_sym.tap { |v| in_validate!(v, stages) } }
      end

      option :build_dir,
             long: "--build-dir PATH",
             description: "Directory where build cache stored ($HOME/.dapp/builds/<dapp name> by default)."

      def run(argv = ARGV)
        self.class.parse_options(self, argv)
        run_dapp_command(run_method, options: cli_options(dimgs_patterns: cli_arguments))
      end
    end
  end
end
