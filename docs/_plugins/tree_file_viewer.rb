require_relative 'plugin_utils'

module Jekyll
  module TreeFileViewer
    class TreeFileViewerTag < Liquid::Tag

      def initialize(tag_name, params_as_string, tokens)
        super
        @params_as_string = params_as_string
      end

      def render(context)
        result = ""

        begin
          @unnamed_params, @named_params = Utils.parse_params(context, @params_as_string)
          Utils.validate_params(@unnamed_params, @named_params, {
            unnamed: [
              {}
            ],
            named: [
              { name: "default_file" }
            ]
          })

          @rel_tree_root = "/" + @unnamed_params[0].delete_prefix("/").delete_suffix("/")
          if @named_params["default_file"]
            @default_active_file_path = "/" + @named_params["default_file"].delete_prefix("/")
          end

          @static_files_root = context.registers[:site].config['source']
          @tree_root = Pathname.new(File.join(@static_files_root, @rel_tree_root))

          file_paths = []
          context.registers[:site].static_files.each do |file|
            if file.path.start_with?(@tree_root.to_s)
              file_paths.push(Pathname.new(file.path))
            end
          end

          files_dirs = [@tree_root]
          file_paths.each do |path|
            unless files_dirs.include?(path.dirname)
              files_dirs.push(path.dirname)
            end

            path.descend do |sub_path|
              if !sub_path.to_s.start_with?(@tree_root.to_s) or files_dirs.include?(sub_path) or !sub_path.directory?
                next
              end

              files_dirs.push(sub_path)
            end
          end

          result = %Q(
<div class="viewer__wrap">
<div class="directory-structure" markdown="0">
)

          result += build_directory_structure(file_paths, files_dirs, @tree_root)

          result += %Q(
</div>
<div class="files-view__wrap">
)

          result += build_files_contents(file_paths)
        rescue => e
          Jekyll.logger.abort_with("[tree_file_viewer] FATAL:", e.message)
        end

        result + %Q(
</div>
</div>
)
      end

      def build_directory_structure(files, dirs, root)
        tree_root_depth = @tree_root.to_s.split("/").length
        root_depth = root.to_s.split("/").length
        is_tree_root_level = root_depth == tree_root_depth

        child_dirs = dirs.select do |dir|
          dir_depth = dir.to_s.split("/").length
          dir.to_s.start_with?(root.to_s) and dir_depth == root_depth + 1
        end

        result = ""

        unless is_tree_root_level
          result += %Q(
<div class="folder__wrap #{root_depth > tree_root_depth + 1 ? "hidden child" : nil}" data-depth="#{root_depth - tree_root_depth}">
<div class="folder">
<span class="folder-icon"></span>
<span class="folder-name">#{root.basename}</span>
</div>
)
        end

        child_dirs.each do |dir|
          result += build_directory_structure(files, dirs, dir)
        end

        files.each_with_index do |file,index|
          unless file.dirname.eql?(root)
            next
          end

          result += %Q(
<div class="file__wrap #{is_file_active(file, index) ? "active" : nil} #{is_tree_root_level ? nil : "hidden child"}">
<span class="file-icon"></span>
<div class="file-name" data-file-name="#{file.relative_path_from(@static_files_root).to_s}">#{file.basename}</div>
</div>
)
        end

        unless is_tree_root_level
          result += %Q(
</div>
)
        end

        result
      end

      def build_files_contents(files)
        result = ""
        files.each_with_index do |file,index|
          result += %Q(
<div class="file-view #{is_file_active(file, index) ? "active" : nil}" data-file-view="#{file.relative_path_from(@static_files_root).to_s}" markdown="1">```#{file.extname.delete_prefix(".")}
#{File.read(file)}
```
</div>
)
        end
        result
      end

      def is_file_active(file, index)
        if @default_active_file_path != nil
          if @default_active_file_path == "/" + file.relative_path_from(@tree_root).to_s
            return true
          end
        else
          index == 0
        end
      end
    end
  end
end

Liquid::Template.register_tag('tree_file_viewer', Jekyll::TreeFileViewer::TreeFileViewerTag)