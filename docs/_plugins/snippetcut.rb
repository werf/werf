module Jekyll
  module SnippetCut
    class SnippetCutTag < Liquid::Block
      @@DEFAULTS = {
          :name => 'myfile.yaml',
          :url => '/asdasda/myfile.yaml'
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
<div class="expand">
<p><strong>#{@config[:name]}</strong> <a href="#{@config[:url]}">🔗</a></p>
#{rendered_content}
</div>
        )
      end
    end
  end
end

Liquid::Template.register_tag('snippetcut', Jekyll::SnippetCut::SnippetCutTag)
