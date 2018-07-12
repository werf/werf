module Dapp
  class Dapp
    module Slug
      SLUG_SEPARATOR = '-'.freeze
      SLUG_V2_LIMIT_LENGTH = 53

      def consistent_uniq_slugify(s)
        return s unless should_be_slugged?(s)
        consistent_uniq_slug_reg =~ s.tr('/', '-').slugify.squeeze('--')
        consistent_uniq_slug = Regexp.last_match(1)
        murmur_hash = MurmurHash3::V32.str_hexdigest(s)
        [].tap do |slug|
          slug << begin
            unless consistent_uniq_slug.nil?
              index = ENV['DAPP_SLUG_V2'] ? SLUG_V2_LIMIT_LENGTH - murmur_hash.length - SLUG_SEPARATOR.length - 1 : -1
              consistent_uniq_slug[0..index]
            end
          end
          slug << murmur_hash
        end.compact.join(SLUG_SEPARATOR)
      end

      def should_be_slugged?(s)
        !(/^#{consistent_uniq_slug_reg}$/ =~ s)
      end

      def consistent_uniq_slug_reg
        /(?!-)((-?[a-z0-9]+)+)(?<!-)/
      end
    end # Slug
  end # Dapp
end # Dapp
