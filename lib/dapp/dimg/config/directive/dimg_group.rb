module Dapp
  module Dimg
    module Config
      module Directive
        class DimgGroup < Base
          include DimgGroupBase
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

          def git(url = nil, &blk)
            check_dimg_directive_order(:git)
            super
          end

          def mount(to, &blk)
            check_dimg_directive_order(:mount)
            super
          end

          protected

          def before_dimg_eval(dimg)
            before_eval_base(dimg)
          end

          def before_dimg_group_eval(dimg_group)
            before_eval_base(dimg_group)
          end

          def before_eval_base(obj)
            pass_to(obj)
            pass_context_artifact_groups_to(obj)
          end

          def pass_context_artifact_groups_to(obj)
            obj.instance_variable_set(:@_artifact_groups, [obj._artifact_groups, clone_variable(_context_artifact_groups)].flatten)
          end

          def check_dimg_directive_order(directive)
            dapp.log_config_warning(desc: { code: 'wrong_using_base_directive',
                                            data: { directive: directive },
                                            context: 'warning' }) if _dimg.any? || _dimg_group.any? || _artifact.any?
          end

          def check_dimg_group_directive_order(directive)
            dapp.log_config_warning(desc: { code: 'wrong_using_directive',
                                            data: { directive: directive },
                                            context: 'warning' }) if _dimg.any? || _dimg_group.any?
          end
        end
      end
    end
  end
end
