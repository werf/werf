# Outputs the reading time

# Read this in “about 4 minutes”
# Put into your _plugins dir in your Jekyll site
# Usage: Read this in about {{ page.content | reading_time }}

module ReadingTimeFilter
  def reading_time( input )
    words_per_minute = 180

    words = input.split.size;
    minutes = ( words / words_per_minute ).floor
    minutes_label = minutes === 1 ? " minute" : " minutes"
    minutes > 0 ? "about #{minutes} #{minutes_label}" : "less than 1 minute"
  end
end

Liquid::Template.register_filter(ReadingTimeFilter)




module Jekyll
  module SWFObject
    class SWFObjectTag < Liquid::Block
      @@DEFAULTS = {
          :title => 'Подробности',
      }

      def self.DEFAULTS
        return @@DEFAULTS
      end

      def initialize(tag_name, markup, tokens)
        super

        @config = {}
        # set defaults
        override_config(@@DEFAULTS)

        params = markup.scan /([a-z]+)\=\"(.+?)\"/
        if params.size > 0
          config = {}
          params.each do |param|
            config[param[0].to_sym] = param[1]
          end
          puts config
          override_config(config)
        end
      end

      def override_config(config)
        config.each { |key,value| @config[key] = value }
        puts @config
      end

      def render(context)
        content = super

        rendered_content = Jekyll::Converters::Markdown::KramdownParser.new(Jekyll.configuration()).convert(content)

        <<-HTML.gsub /^\s+/, '' # remove whitespaces from heredocs
        <div class="details">
            <p class="details__lnk"><a href="javascript:void(0)" class="details__summary">#{@config[:title]}</a></p>
            <div class="details__content" markdown="1">
                <div class="expand">
                    #{rendered_content}
                </div>
            </div>
        </div>
        HTML
      end
    end
  end
end

Liquid::Template.register_tag('offtopic', Jekyll::SWFObject::SWFObjectTag)




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
