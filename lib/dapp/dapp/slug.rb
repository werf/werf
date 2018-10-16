module Dapp
  class Dapp
    module Slug
      def consistent_uniq_slugify(s)
        ruby2go_slug({ options: { data: s } }).tap do |res|
          raise Error::Build, code: :ruby2go_slug_failed_unexpected_error, data: { message: res["error"] } unless res["error"].nil?
          break res['data']
        end
      end
    end # Slug
  end # Dapp
end # Dapp
