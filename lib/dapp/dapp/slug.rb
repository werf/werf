module Dapp
  class Dapp
    module Slug
      def consistent_uniq_slugify(s)
        if should_be_slugged?(s)
          consistent_uniq_slug_reg =~ s.gsub("/", "-").slugify.squeeze('--')
          [].tap do |slug|
            slug << Regexp.last_match(1) unless Regexp.last_match(1).nil?
            slug << MurmurHash3::V32.str_hexdigest(s)
          end.compact.join('-')
        else
          s
        end
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
