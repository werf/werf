# frozen_string_literal: true

module Jekyll
  module Utils
    def parse_params(context, params_as_string)
      templated_params_as_string = Liquid::Template
                                     .parse(params_as_string)
                                     .render(context)
                                     .gsub(%r!\\\{\\\{|\\\{\\%!, '\{\{' => "{{", '\{\%' => "{%")

      params = templated_params_as_string.scan(/(?:"(?:\\.|[^"])*"|[^" ])+/)

      unnamed_params = params.select { |param| !param.include?("=") }.map do |param|
        param = param.strip
        if (param.start_with?('"') and param.end_with?('"')) or (param.start_with?("'") and param.end_with?("'"))
          param = param[1...-1]
        end
        param.strip
      end

      named_params = params.select { |param| param.include?("=") }.to_h do |param|
        parts = param.split("=", 2)
        key = parts[0].strip
        value = parts[1].strip
        if (value.start_with?('"') and value.end_with?('"')) or (value.start_with?("'") and value.end_with?("'"))
          value = value[1...-1]
        end
        value = value.strip
        [key, value]
      end

      [unnamed_params, named_params]
    end

    def validate_params(unnamed, named, scheme)
      validate_unnamed_params(unnamed, scheme)
      validate_named_params(named, scheme)
    end

    private

    def validate_unnamed_params(unnamed, scheme)
      if unnamed == [] or !scheme.key?("unnamed")
        return
      end

      if unnamed.length > scheme[:unnamed].length
        raise ArgumentError.new("Too many unnamed parameters. Expected #{scheme[:unnamed].length}, got #{unnamed.length}")
      end

      if unnamed.length < scheme[:unnamed].length
        raise ArgumentError.new("Too few unnamed parameters. Expected #{scheme[:unnamed].length}, got #{unnamed.length}")
      end

      unnamed.each do |param|
        if param == ""
          raise ArgumentError.new("Cannot have empty unnamed parameters")
        end

        scheme_param = scheme[:unnamed][unnamed.index(param)]
        if scheme_param[:regex] and !scheme_param[:regex].match?(param)
          raise ArgumentError.new("Invalid unnamed parameter \"#{param}\", must match #{param[:regex]}")
        end
      end
    end

    def validate_named_params(named, scheme)
      if named == {} or !scheme.key?("named")
        return
      end

      if named.keys.length > scheme[:named].length
        raise ArgumentError.new("Too many named parameters. Expected at most #{scheme[:named].length}, got #{named.keys.length}")
      end

      required_named_params = scheme[:named].select { |param| param[:required] }
      required_named_params.each do |param|
        unless named.keys.include?(param[:name])
          raise ArgumentError.new("Missing required named parameter \"#{param[:name]}\"")
        end
      end

      named.each do |key, value|
        if key == ""
          raise ArgumentError.new("Cannot have empty keys in named parameters")
        end

        if value == ""
          raise ArgumentError.new("Cannot have empty values in named parameters")
        end

        scheme_param = scheme[:named].find { |param| param[:name] == key }
        unless scheme_param
          raise ArgumentError.new("Invalid named parameter \"#{key}\"")
        end

        if scheme_param[:regex] and !scheme_param[:regex].match?(value)
          raise ArgumentError.new("Invalid value \"#{value}\" for named parameter \"#{key}\"")
        end
      end
    end
  end
end