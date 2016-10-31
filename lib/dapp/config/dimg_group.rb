module Dapp
  module Config
    class DimgGroup < DimgGroupBase
      include Dimg::InstanceMethods

      def dimg(_name = nil, &_blk)
        super.tap do |dimg|
          pass_to_default(dimg)
        end
      end

      def chef(&blk)
        check_dimg_directive_order
        super
      end

      def shell(&blk)
        check_dimg_directive_order
        super
      end

      def docker(&blk)
        check_dimg_directive_order
        super
      end

      def artifact(&blk)
        check_dimg_group_directive_order
        super
      end

      def git_artifact(type_or_git_repo, &blk)
        check_dimg_directive_order
        super
      end

      def mount(to, &blk)
        check_dimg_directive_order
        super
      end

      protected

      def pass_to_default(dimg)
        pass_to_custom(dimg, :clone)
      end

      def pass_to_custom(obj, clone_method)
        passing_directives.each do |directive|
          directive_value = if (variable = instance_variable_get(directive)).is_a? Directive::Base
                              variable.send(clone_method)
                            else
                              marshal_dup(variable)
                            end
          obj.instance_variable_set(directive, directive_value) unless directive_value.nil?
        end
        obj
      end

      def passing_directives
        [:@_chef, :@_shell, :@_docker, :@_git_artifact, :@_mount, :@_artifact]
      end

      def check_dimg_directive_order
        raise if check_dimg_group_directive_order or _artifact.any? # TODO: adding order
      end

      def check_dimg_group_directive_order
        raise if _dimg.any? or _dimg_group.any? # TODO: adding order
      end
    end
  end
end
