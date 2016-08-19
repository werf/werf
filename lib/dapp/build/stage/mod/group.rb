module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Group
        module Group
          def image_build
            log_group_name if group_closed?
            application.with_log_indent { super }
          end

          def log_group_name
            application.log_step(application.t(code: group_name, context: :group))
          end

          def group_closed?
            !group_opened?
          end

          def group_opened?
            return false if prev_group_stage.nil?
            prev_group_stage.group_opened? || image_should_be_build?
          end

          def prev_group_stage
            prev_stage if prev_stage.respond_to?(:group_name) && prev_stage.group_name == group_name
          end

          def group_name
            class_to_lowercase(self.class.name.split('::')[-2])
          end

          def name_context
            [super, group_name].join('.')
          end
        end
      end # Mod
    end # Stage
  end # Build
end # Dapp
