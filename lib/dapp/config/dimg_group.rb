module Dapp
  module Config
    # DimgGroup
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

      def git(type_or_git_repo, &blk)
        check_dimg_directive_order(:git)
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

      def check_dimg_directive_order(directive)
        project.log_config_warning(desc: { code: 'wrong_using_base_directive',
                                           data: { directive: directive },
                                           context: 'warning' }) if _dimg.any? || _dimg_group.any? || _artifact.any?
      end

      def check_dimg_group_directive_order(directive)
        project.log_config_warning(desc: { code: 'wrong_using_directive',
                                           data: { directive: directive },
                                           context: 'warning' }) if _dimg.any? || _dimg_group.any?
      end
    end
  end
end
