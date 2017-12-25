module Dapp
  module Dimg
    module Image
      class Dimg < Stage
        def export!(export_name = name)
          super(export_name)
        end

        def tag!(tag_name = name)
          super(tag_name)
        end
      end
    end # Image
  end # Dimg
end # Dapp
