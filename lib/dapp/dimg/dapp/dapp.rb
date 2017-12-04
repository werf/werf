module Dapp
  module Dimg
    module Dapp
      module Dapp
        include Dappfile

        include Command::Common
        include Command::Run
        include Command::Build
        include Command::Bp
        include Command::Push
        include Command::Spush
        include Command::Tag
        include Command::List
        include Command::Stages::CleanupLocal
        include Command::Stages::CleanupRepo
        include Command::Stages::FlushLocal
        include Command::Stages::FlushRepo
        include Command::Stages::Push
        include Command::Stages::Pull
        include Command::Stages::Common
        include Command::CleanupRepo
        include Command::FlushRepo
        include Command::Cleanup
        include Command::Mrproper
        include Command::StageImage
        include Command::BuildContext::Import
        include Command::BuildContext::Export
        include Command::BuildContext::Common
      end
    end
  end
end

::Dapp::Dapp.send(:include, ::Dapp::Dimg::Dapp::Dapp)
