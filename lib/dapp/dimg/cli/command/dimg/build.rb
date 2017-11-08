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
        introspected_stages = [
          :from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install,
          :g_a_post_install_patch, :after_install_artifact, :before_setup, :before_setup_artifact,
          :g_a_pre_setup_patch, :setup, :g_a_post_setup_patch, :after_setup_artifact, :g_a_latest_patch, :docker_instructions
        ]
        artifact_introspected_stages = [
          :from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install,
          :g_a_post_install_patch, :after_install_artifact, :before_setup, :before_setup_artifact,
          :g_a_pre_setup_patch, :setup, :after_setup_artifact, :g_a_artifact_patch, :build_artifact
        ]

        option :tmp_dir_prefix,
               long: '--tmp-dir-prefix PREFIX',
               description: 'Tmp directory prefix (/tmp by default). Used for build process service directories.'

        option :lock_timeout,
               long: '--lock-timeout TIMEOUT',
               description: 'Redefine resource locking timeout (in seconds)',
               proc: ->(v) { v.to_i }

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
               description: "Introspect one of the following stages (#{list_msg_format(introspected_stages)})",
               proc: proc { |val| val.to_sym.tap { |v| in_validate!(v, introspected_stages) } }

        option :introspect_artifact_stage,
               long: '--introspect-artifact-stage STAGE',
               description: "Introspect one of the following stages (#{list_msg_format(artifact_introspected_stages)})",
               proc: proc { |val| val.to_sym.tap { |v| in_validate!(v, artifact_introspected_stages) } }

        option :ssh_key,
               long: '--ssh-key SSH_KEY',
               description: ['Enable only specified ssh keys ',
                             '(use system ssh-agent by default)'].join,
               default: nil,
               proc: ->(v) { composite_options(:ssh_key) << v }

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
