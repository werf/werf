module Dapp
  class Dapp
    module Command
      module Example
        module Create
          def example_create(example_name)
            example_exist!(example_name)

            _example_repo_blobs_entries(example_name).each do |root, entry|
              file_path = _example_slice_cwd(example_name, File.join(root, entry[:name]))
              content = _example_git_repo.lookup_object(entry[:oid]).content

              begin
                FileUtils.mkdir_p(File.dirname(file_path))
                if entry[:filemode] == 0o120000 # symlink
                  FileUtils.symlink(content, file_path)
                else
                  IO.write(file_path, content)
                  FileUtils.chmod(entry[:filemode], file_path)
                end
              rescue Errno::EEXIST => e
                log_warning("File `#{file_path}` skipped: `#{e.message}`")
              end
            end
          end

          def example_exist!(example_name)
            return if example_exist?(example_name)
            raise Error::Command, code: :example_not_exist, data: { name: example_name, url: _examples_git_repo_url, path: _examples_dir }
          end

          def example_exist?(example_name)
            _example_list.include?(example_name)
          end

          def _example_repo_blobs_entries(example_name)
            _example_git_repo
              .blobs_entries(_example_git_repo.latest_commit(_examples_git_repo_branch), paths: [_example_directory(example_name)])
              .reject { |_, entry| entry[:filemode] == 0o160000 }
          end

          def _example_directory(example_name)
            File.expand_path(File.join('/', _examples_dir, example_name))[1..-1]
          end

          def _example_slice_cwd(example_name, path)
            path
              .reverse
              .chomp(_example_directory(example_name).reverse)
              .chomp('/')
              .reverse
          end
        end
      end
    end
  end
end
