module Dapp::Dimg::CLI
  module Options
    module Introspection
      def self.extended(klass)
        artifact_stages = [
          :from, :before_install, :before_install_artifact, :g_a_archive, :g_a_pre_install_patch, :install,
          :g_a_post_install_patch, :after_install_artifact, :before_setup, :before_setup_artifact,
          :g_a_pre_setup_patch, :setup, :after_setup_artifact, :g_a_artifact_patch, :build_artifact
        ]

        klass.class_eval do
          option :introspect_error,
                 long: '--introspect-error',
                 boolean: true,
                 default: false

          option :introspect_before_error,
                 long: '--introspect-before-error',
                 boolean: true,
                 default: false

          option :introspect_stage,
                 long:        '--introspect-stage STAGE',
                 description: "Introspect one of the following stages (#{list_msg_format(klass.const_get(:DIMG_STAGES))})",
                 proc:        klass.const_get(:STAGE_PROC).call(klass.const_get(:DIMG_STAGES))

          option :introspect_before,
                 long:        '--introspect-before STAGE',
                 description: "Introspect stage before one of the following stages (#{list_msg_format(klass.const_get(:DIMG_STAGES))})",
                 proc:        klass.const_get(:STAGE_PROC).call(klass.const_get(:DIMG_STAGES))

          option :introspect_artifact_stage,
                 long:        '--introspect-artifact-stage STAGE',
                 description: "Introspect one of the following stages (#{list_msg_format(artifact_stages)})",
                 proc:        klass.const_get(:STAGE_PROC).call(artifact_stages)

          option :introspect_artifact_before,
                 long:        '--introspect-artifact-before STAGE',
                 description: "Introspect stage before one of the following stages (#{list_msg_format(artifact_stages)})",
                 proc:        klass.const_get(:STAGE_PROC).call(artifact_stages)
        end
      end
    end
  end
end
