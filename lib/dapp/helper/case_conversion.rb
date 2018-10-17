module Dapp
  module Helper
    module CaseConversion
      class << self
        def snake_case_to_camel_case(value)
          value.to_s.split('_').collect(&:capitalize).join
        end

        def snake_case_to_lower_camel_case(value)
          res = snake_case_to_camel_case(value)
          res[0] = res[0].downcase
          res
        end
      end # << self
    end # CaseConversion
  end # Helper
end # Dapp
