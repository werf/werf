module Dapp::Builder
  # Локальный cookbook сборщик с правилами
  # сборки dimg образов для chef сборщика.
  class Chef::Cookbook
    CHECKSUM_PATTERNS = %w(
      attributes/**/*
      recipes/**/*
      files/**/*
      templates/**/*
    ).freeze

    attr_reader :builder

    attr_reader :path
    attr_reader :vendor_path
    attr_reader :berksfile
    attr_reader :metadata
    attr_reader :enabled_recipes
    attr_reader :enabled_modules

    def initialize(builder, path:, berksfile:, metadata:, enabled_recipes:, enabled_modules:)
      @builder = builder
      @path = Pathname.new(path)
      @berksfile = berksfile
      @metadata = metadata
      @enabled_recipes = enabled_recipes
      @enabled_modules = enabled_modules
    end

    def name
      metadata.name
    end

    def local_paths
      @local_paths ||= [].tap do |paths|
        paths << path
        paths.push *berksfile.local_cookbooks.values.map {|cookbook| cookbook[:path]}
        paths.uniq!
      end
    end

    # TODO Сookbook'и зависимости на локальной файловой системе
    # TODO должны обрабатываться рекурсивно. Чтобы контрольная сумма
    # TODO учитывала все зависимости всех зависимых локальных cookbook'ов.
    def checksum
      @checksum ||= begin
        all_local_paths = local_paths
          .product(CHECKSUM_PATTERNS)
          .map {|cb, dir| Dir[cb.join(dir)]}
          .flatten
          .map(&Pathname.method(:new))

        builder.dimg.hashsum [
          builder.dimg.paths_content_hashsum(all_local_paths),
          *all_local_paths.map {|p| p.relative_path_from(self.path).to_s}.sort
        ].compact
      end
    end

    def vendor_path
      builder.dimg.project.lock(_vendor_lock_name, default_timeout: 120) do
        vendor! unless _vendor_path.join('.created_at').exist?
      end
      _vendor_path
    end

    def vendor!
      volumes_from = [builder.dimg.project.base_container, builder.chefdk_container]

      builder.dimg.project.log_secondary_process(builder.dimg.project.t(code: _vendor_process_name)) do
        volumes_from = [builder.dimg.project.base_container, builder.chefdk_container]

        vendor_commands = [
          "#{builder.dimg.project.mkdir_bin} -p ~/.ssh",
          "echo \"Host *\" >> ~/.ssh/config",
          "echo \"    StrictHostKeyChecking no\" >> ~/.ssh/config",
          *local_paths.map {|path| "#{builder.dimg.project.rsync_bin} --archive --relative #{path} /tmp/local_cookbooks"},
          "cd /tmp/local_cookbooks/#{path}",
          '/.dapp/deps/chefdk/bin/berks vendor /tmp/cookbooks',
          ["#{builder.dimg.project.find_bin} /tmp/cookbooks -type d -exec #{builder.dimg.project.bash_bin} -ec '",
           "#{builder.dimg.project.install_bin} -o #{Process.uid} -g #{Process.gid} --mode $(#{builder.dimg.project.stat_bin} -c %a {}) -d ",
           "#{_vendor_path}/$(echo {} | #{builder.dimg.project.sed_bin} -e \"s/^\\/tmp\\/cookbooks//\")' \\;"].join,
          ["#{builder.dimg.project.find_bin} /tmp/cookbooks -type f -exec #{builder.dimg.project.bash_bin} -ec '",
           "#{builder.dimg.project.install_bin} -o #{Process.uid} -g #{Process.gid} --mode $(#{builder.dimg.project.stat_bin} -c %a {}) {} ",
           "#{_vendor_path}/$(echo {} | #{builder.dimg.project.sed_bin} -e \"s/\\/tmp\\/cookbooks//\")' \\;"].join,
          "#{builder.dimg.project.install_bin} -o #{Process.uid} -g #{Process.gid} --mode 0644 <(#{builder.dimg.project.date_bin} +%s.%N) #{_vendor_path.join('.created_at')}"
        ]

        builder.dimg.project.shellout!(
          ['docker run --rm',
           volumes_from.map {|container| "--volumes-from #{container}"}.join(' '),
           *local_paths.map {|path| "--volume #{path}:#{path}"},
           ("--volume #{builder.dimg.project.ssh_auth_sock}:/tmp/dapp-ssh-agent" if builder.dimg.project.ssh_auth_sock),
           "--volume #{_vendor_path.tap(&:mkpath)}:#{_vendor_path}",
           ('--env SSH_AUTH_SOCK=/tmp/dapp-ssh-agent' if builder.dimg.project.ssh_auth_sock),
           "dappdeps/berksdeps:0.1.0 #{builder.dimg.project.bash_bin} -ec '#{builder.dimg.project.shellout_pack(vendor_commands.join(' && '))}'"].compact.join(' '),
          log_verbose: builder.dimg.project.log_verbose?
        )
      end
    end

    def stage_cookbooks_path(stage)
      _stage_cookbooks_path(stage).tap do |_cookbooks_path|
        @stage_cookbooks_installed ||= {}
        @stage_cookbooks_installed[stage] ||= true.tap {install_stage_cookbooks(stage)}
      end
    end

    def stage_enabled_modules(stage)
      @stage_enabled_modules ||= {}
      @stage_enabled_modules[stage] ||= enabled_modules.select {|cookbook| stage_entry_exist?(stage, cookbook, stage)}
    end

    def stage_enabled_recipes(stage)
      @stage_enabled_recipes ||= {}
      @stage_enabled_recipes[stage] ||= enabled_recipes.select {|recipe| stage_entry_exist?(stage, self.name, recipe)}
    end

    def stage_checksum(stage)
      paths = Dir[stage_cookbooks_path(stage).join('**/*')].map(&Pathname.method(:new))

      builder.dimg.hashsum [
        builder.dimg.paths_content_hashsum(paths),
        *paths.map {|p| p.relative_path_from(stage_cookbooks_path(stage)).to_s}.sort,
        stage == :before_install ? builder.chefdk_image : nil
      ].compact
    end

    protected

    def _vendor_path
      builder.dimg.build_path.join('cookbooks', checksum)
    end

    def _vendor_process_name
      "process.vendoring_builder_cookbooks"
    end

    def _vendor_lock_name
      "#{builder.dimg.project.name}.cookbooks.#{checksum}"
    end

    def stage_entry_exist?(stage, cookbook, entrypoint)
      stage_cookbooks_path(stage).join(cookbook, 'recipes', "#{entrypoint}.rb").exist?
    end

    def install_stage_cookbooks(stage)
      _stage_cookbooks_path(stage).mkpath

      stage_install_paths(stage).each do |cookbook_path, paths|
        cookbook = cookbook_path.basename.to_s

        paths.each do |from, to|
          if from.nil?
            to_path = _stage_cookbooks_path(stage).join(cookbook, 'recipes/void.rb')
            to_path.parent.mkpath
            FileUtils.touch to_path
          else
            from_path = cookbook_path.join(from)
            to_path = _stage_cookbooks_path(stage).join(cookbook, to)
            if from_path.directory? && to_path.exist?
              Dir[from_path.join('**/*')]
                .map(&Pathname.method(:new))
                .each do |from_subpath|
                  to_subpath = to_path.join(from_subpath.relative_path_from(from_path))
                  raise(::Dapp::Builder::Chef::Error,
                    code: :stage_path_overlap,
                    data: { stage: stage,
                            cookbook: cookbook,
                            from: from_subpath.relative_path_from(cookbook_path),
                            to: to_subpath.relative_path_from(_stage_cookbooks_path(stage).join(cookbook)) }
                  ) if to_subpath.exist?

                  to_subpath.parent.mkpath
                  FileUtils.cp_r from_subpath, to_subpath
                end
            else
              to_path.parent.mkpath
              FileUtils.cp_r from_path, to_path
            end
          end
        end
      end
    end

    def stage_install_paths(stage)
      select_existing_paths = proc do |cookbook_path, paths|
        paths.select {|from, _| cookbook_path.join(from).exist?}
      end

      common_paths = [['metadata.json', 'metadata.json']]

      install_paths = Dir[vendor_path.join('*')]
                      .map(&Pathname.method(:new))
                      .map do |cookbook_path|
        cookbook_name = File.basename cookbook_path
        is_builder = (cookbook_name == self.name)
        is_dimod = cookbook_name.start_with? 'dimod-'
        dimod_enabled = is_dimod && enabled_modules.include?(cookbook_name)
        paths = nil

        if is_builder
          common_dapp_paths = select_existing_paths.call(
            cookbook_path,
            [
              *common_paths,
              ["files/#{stage}/common", 'files/default'],
              ["templates/#{stage}/common", 'templates/default'],
              *enabled_recipes.flat_map do |recipe|
                [["files/#{stage}/#{recipe}", 'files/default'],
                ["templates/#{stage}/#{recipe}", 'templates/default']]
              end
            ]
          )

          recipe_paths = enabled_recipes
            .map {|recipe| ["recipes/#{stage}/#{recipe}.rb", "recipes/#{recipe}.rb"]}
            .select {|from, _| cookbook_path.join(from).exist?}

          if recipe_paths.any?
            paths = [*recipe_paths, *common_dapp_paths]
          else
            paths = [nil, *common_dapp_paths]
          end
        elsif is_dimod && dimod_enabled
          common_dimod_paths = select_existing_paths.call(
            cookbook_path,
            [
              *common_paths,
              ["files/#{stage}", 'files/default'],
              ['files/common', 'files/default'],
              ["templates/#{stage}", 'templates/default'],
              ['templates/common', 'templates/default'],
              ["attributes/#{stage}.rb", "attributes/#{stage}.rb"],
              ['attributes/common.rb', 'attributes/common.rb']
            ]
          )

          recipe_path = "recipes/#{stage}.rb"
          if cookbook_path.join(recipe_path).exist?
            paths = [[recipe_path, recipe_path], *common_dimod_paths]
          else
            paths = [nil, *common_dimod_paths]
          end
        elsif !is_dimod
          paths = [['.', '.']]
        end

        [cookbook_path, paths] if paths && paths.any?
      end.compact
    end

    def _stage_cookbooks_path(stage)
      builder.stage_build_path(stage).join('cookbooks')
    end
  end # Chef::Cookbook
end # Dapp::Builder
