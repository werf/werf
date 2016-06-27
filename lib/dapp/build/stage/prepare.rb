module Dapp
  module Build
    module Stage
      class Prepare < Base
        include From::Centos7
        include From::Ubuntu1404
        include From::Ubuntu1604

        def signature
          send(:"#{constructor_method_prefix}_signature")
        end

        def image
          @image ||= send(:"#{constructor_method_prefix}_image")
        end

        private

        def constructor_method_prefix
          "#{application.conf[:from].to_s.split(/[:.]/).join}"
        end
      end # Prepare
    end # Stage
  end # Build
end # Dapp
