module Dapp
  module Dimg
    module Dapp
      module Command
        module Push
          def push
            #require 'ruby-prof'
            #RubyProf.start
            log_step_with_indent(:stages) { stages_push } if with_stages?
            dimg_import_export_base do |dimg|
              dimg.export!(option_repo, format: push_format(dimg.config._name))
            end
            # FIXME: rework images cache, then profile
            #result = RubyProf.stop
            #printer = RubyProf::MultiPrinter.new(result)
            #printer.print(path: '/tmp/testdapp.push.profile', profile: 'profile')
          end
        end
      end
    end
  end # Dimg
end # Dapp
