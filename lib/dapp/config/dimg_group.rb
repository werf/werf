module Dapp
  module Config
    class DimgGroup < DimgGroupBase
      include Dimg::InstanceMethods

      def chef(&blk)
        check_dimg_directive_order(:chef)
        super
      end

      def shell(&blk)
        check_dimg_directive_order(:shell)
        super
      end

      def docker(&blk)
        check_dimg_directive_order(:docker)
        super
      end

      def artifact(&blk)
        check_dimg_group_directive_order(:artifact)
        super
      end

      def git_artifact(type_or_git_repo, &blk)
        check_dimg_directive_order(:git_artifact)
        super
      end

      def mount(to, &blk)
        check_dimg_directive_order(:mount)
        super
      end

      protected

      def before_dimg_eval(dimg)
        pass_to_default(dimg)
      end

      def pass_to_default(dimg)
        pass_to_custom(dimg, :clone)
      end

      def pass_to_custom(obj, clone_method)
        passing_directives.each do |directive|
          next if (variable = instance_variable_get(directive)).nil?
          obj.instance_variable_set(directive, variable.send(clone_method))
        end
        obj.instance_variable_set(:@_artifact, _artifact)
        obj.instance_variable_set(:@_builder, _builder)
        obj
      end

      def passing_directives
        [:@_chef, :@_shell, :@_docker, :@_git_artifact, :@_mount]
      end

      def check_dimg_directive_order(directive)
        project.log_config_warning(desc: { code: 'wrong_using_base_directive',
                                           data: { directive: directive },
                                           context: 'warning' }) if _dimg.any? or _dimg_group.any? or _artifact.any?
      end

      def check_dimg_group_directive_order(directive)
        project.log_config_warning(desc: { code: 'wrong_using_directive',
                                           data: { directive: directive },
                                           context: 'warning' }) if _dimg.any? or _dimg_group.any?
      end
    end
  end
end
