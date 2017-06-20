module Dapp
  module Dimg
    module Build
      module Stage
        module Mod
          module Group
            def log_image_build
              log_group_name if group_should_be_opened?
              dimg.dapp.with_log_indent { super }
            end

            def log_group_name
              dimg.dapp.log_step(dimg.dapp.t(code: group_name, context: :group))
            end

            def group_name
              class_to_lowercase(self.class.name.split('::')[-2])
            end

            def group_should_be_opened?
              @group_should_be_opened ||= begin
                if empty?
                  false
                elsif prev_group_stage.nil?
                  true
                else
                  prev_group_stages.all? { |s| !s.group_should_be_opened? }
                end
              end
            end

            def prev_group_stage
              prev_stage if prev_stage.respond_to?(:group_name) && prev_stage.group_name == group_name
            end

            def prev_group_stages
              @prev_group_stages ||= begin
                if prev_group_stage.nil?
                  []
                else
                  prev_group_stage.prev_group_stages + [prev_group_stage]
                end
              end
            end

            def log_name_context
              [super, group_name].join('.')
            end
          end
        end # Mod
      end # Stage
    end # Build
  end # Dimg
end # Dapp
