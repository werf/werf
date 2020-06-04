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

        params = markup.split
        if params.size > 0
          config = {}
          params.each do |param|
            param = param.gsub /\s+/, ''
            key, value = param.split(':',2)
            config[key.to_sym] = value
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

        <<-HTML.gsub /^\s+/, '' # remove whitespaces from heredocs
        <div class="expand">
            <p><strong>#{@config[:name]}</strong> <a href="#{@config[:url]}">🔗</a></p>
            #{rendered_content}
        </div>
        HTML
      end
    end
  end
end

Liquid::Template.register_tag('snippetcut', Jekyll::SnippetCut::SnippetCutTag)
