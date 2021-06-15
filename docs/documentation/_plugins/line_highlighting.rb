# Monkey-patch to allow highlighting lines
module Jekyll
    module Tags
        class HighlightBlock
            def render_rouge(code)
                require "rouge"
                formatter = Rouge::Formatters::HTMLLineHighlighter.new(
                    ::Rouge::Formatters::HTML.new,
                    highlight_lines: parse_highlighted_lines(@highlight_options[:highlight_lines])
                )
                lexer = ::Rouge::Lexer.find_fancy(@lang, code) || Rouge::Lexers::PlainText
                formatter.format(lexer.lex(code))
            end

            private

            def parse_highlighted_lines(lines_string)
                return [] if lines_string.nil?

                lines_string.map(&:to_i)
            end
        end
    end
end