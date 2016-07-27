module SpecHelper
  module Expect
    def expect_exception_code(code:)
      expect { yield }.to raise_error { |error| expect(error.net_status[:code]).to be(code) }
    end
  end
end
