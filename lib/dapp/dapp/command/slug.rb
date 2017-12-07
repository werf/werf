module Dapp
  class Dapp
    module Command
      module Slug
        def slug(str)
          puts consistent_uniq_slugify(str)
        end
      end
    end
  end
end
