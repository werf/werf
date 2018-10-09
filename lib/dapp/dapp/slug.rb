module Dapp
  class Dapp
    module Slug
      SLUG_SEPARATOR = '-'.freeze
      SLUG_V2_LIMIT_LENGTH = 53

      def self.included(base)
        if ENV['DAPP_SLUG_V3']
          base.include(V3)
        else
          base.include(V1V2)
        end
      end

      module V1V2
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
      end

      module V3
        def consistent_uniq_slugify(s)
          return s unless should_be_slugged?(s)
          slug = consistent_uniq_slug(s)
          murmur_hash = MurmurHash3::V32.str_hexdigest(s)
          [].tap do |res|
            res << begin
              unless slug.empty?
                index = SLUG_V2_LIMIT_LENGTH - murmur_hash.length - SLUG_SEPARATOR.length - 1
                slug[0..index]
              end
            end
            res << murmur_hash
          end.compact.join(SLUG_SEPARATOR)
        end

        def should_be_slugged?(s)
          consistent_uniq_slug(s) != s || s.length > SLUG_V2_LIMIT_LENGTH
        end

        def consistent_uniq_slug(s)
          ''.tap do |res|
            status = :empty
            s.to_s.chars.each do |ch|
              next if (s_ch = ch.slugify).empty?

              if s_ch !~ /[[:punct:]|[:blank:]]/
                res << s_ch
                status = :non_empty if status == :empty
              elsif status == :non_empty && res[-1] != '-'
                res << '-'
              end
            end
          end.chomp('-')
        end
      end
    end # Slug
  end # Dapp
end # Dapp
