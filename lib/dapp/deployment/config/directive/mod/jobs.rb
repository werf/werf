module Dapp
  module Deployment
    module Config
      module Directive
        module Mod
          module Jobs
            attr_reader :_bootstrap, :_before_apply_job

            def bootstrap(&blk)
              directive_eval(_bootstrap, &blk)
            end

            def before_apply_job(&blk)
              directive_eval(_before_apply_job, &blk)
            end

            def jobs_init_variables!
              @_bootstrap        = Directive::Job.new(dapp: dapp)
              @_before_apply_job = Directive::Job.new(dapp: dapp)
            end
          end
        end
      end
    end
  end
end
