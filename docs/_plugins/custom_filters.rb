module Jekyll
  module CustomFilters
    def true_relative_url(path, page_path)
        # remove first slash if exist
        page_path_relative = page_path.gsub(%r!^/!, "")
        page_depth = page_path_relative.scan(%r!/!).count
        prefix = ""
        page_depth.times{ prefix = prefix + "../" }
        prefix + path.sub(%r!^/!, "")
    end
  end
end

Liquid::Template.register_filter(Jekyll::CustomFilters)
