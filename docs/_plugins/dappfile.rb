module Jekyll
  module Dappfile
    def dappfile(input, base_id)
      input.gsub!(/\<.*?>/) do |match|
        id = [base_id, 'dappfile', match.gsub(/&lt;|&gt;/, ''), rand(1000)].join('-')
        "<span id=\"#{slugify(id)}\">#{escape(match)}</span>"
      end
      newline_to_br(input)
    end
  end
end

Liquid::Template.register_filter(Jekyll::Dappfile)