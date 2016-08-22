module Dapp
  module Build
    module Stage
      # Mod
      module Mod
        # Group
        module Group
          def log_image_build
            return super if should_be_quiet?
            log_group_name if group_should_be_opened?
            application.project.with_log_indent { super }
          end

          def log_group_name
            application.project.log_step(application.project.t(code: group_name, context: :group))
          end

          def group_should_be_opened?
            return image_should_be_build? if prev_group_stage.nil?
            !group_opened? && image_should_be_build?
          end

          def group_opened?
            return image_should_be_build? if prev_group_stage.nil?
            prev_group_stage.group_opened? || prev_group_stage.image_should_be_build?
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
