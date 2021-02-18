module Jekyll
  module CustomFilters
    def true_relative_url(path)
        if !path.instance_of? String
            raise "true_relative_url filter failed: unexpected argument #{path}"
        end

        # remove first slash if exist
        page_path_relative = @context.registers[:page]["url"].gsub(%r!^/!, "")
        page_depth = page_path_relative.scan(%r!/!).count
        prefix = ""
        page_depth.times{ prefix = prefix + "../" }
        prefix + path.sub(%r!^/!, "")
    end

    # get_lang_field_or_raise_error filter returns a field from argument hash
    # returns nil if hash is empty
    # returns hash[page.lang] if hash has the field
    # returns hash["all"] if hash has the field
    # otherwise, raise an error
    def get_lang_field_or_raise_error(hash)
        if !(hash == nil or hash.instance_of? Hash)
            raise "get_lang_field_or_raise_error filter failed: unexpected argument '#{hash}'"
        end

        if hash == nil or hash.length == 0
            return
        end

        lang = @context.registers[:page]["lang"]
        if hash.has_key?(lang)
            return hash[lang]
        elsif hash.has_key?("all")
            return hash["all"]
        else
            raise "get_lang_field_or_raise_error filter failed: the argument '#{hash}' does not have '#{lang}' or 'all' field"
        end
    end
  end
end

Liquid::Template.register_filter(Jekyll::CustomFilters)
