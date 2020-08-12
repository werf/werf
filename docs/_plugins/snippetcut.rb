module Jekyll
  module SnippetCut
    class SnippetCutTag < Liquid::Block
      @@DEFAULTS = {
          :name => 'myfile.yaml',
          :url => '/asdasda/myfile.yaml',
          :limited => false
      }

      def self.DEFAULTS
        return @@DEFAULTS
      end

      def initialize(tag_name, markup, tokens)
        super

        @config = {}
        override_config(@@DEFAULTS)

        params = markup.scan /([a-z]+)\=\"(.+?)\"/
        if params.size > 0
          config = {}
          params.each do |param|
            config[param[0].to_sym] = param[1]
          end
          override_config(config)
        end
      end

      def override_config(config)
        config.each{ |key,value| @config[key] = value }
      end

      def render(context)
        content = super

        rendered_content = Jekyll::Converters::Markdown::KramdownParser.new(Jekyll.configuration()).convert(content)

        %Q(
<div class="snippetcut#{@config[:limited] ? ' snippetcut_limited' : ''}" data-snippetcut>
<div class="snippetcut__title">
  #{if (@config[:url]!='#') then "<a href=\""+@config[:url]+"\" target=\"_blank\" class=\"snippetcut__title-name\" data-snippetcut-name>" else "<span class=\"snippetcut__title-name-text\">" end}
  #{@config[:name]}
  #{if (@config[:url]!='#') then "</a>" else "</span>" end}
<a href="javascript:void(0)" class="snippetcut__title-btn" data-snippetcut-btn-name>копировать имя</a>
<a href="javascript:void(0)" class="snippetcut__title-btn" data-snippetcut-btn-text>копировать текст</a>
</div>
#{rendered_content}
<div class="snippetcut__raw" data-snippetcut-text>#{content}</div>
</div>
        )
  end
        end
        end
        end

        Liquid::Template.register_tag('snippetcut', Jekyll::SnippetCut::SnippetCutTag)
