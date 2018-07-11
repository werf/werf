module Dapp
  class Dapp
    module Command
      module Sample
        module Create
          def sample_create(sample_name)
            sample_exist!(sample_name)

            _sample_repo_blobs_entries(sample_name).each do |root, entry|
              file_path = _sample_slice_cwd(sample_name, File.join(root, entry[:name]))
              content = _sample_git_repo.lookup_object(entry[:oid]).content

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

          def sample_exist!(sample_name)
            return if sample_exist?(sample_name)
            raise Error::Command, code: :sample_not_exist, data: { name: sample_name, url: _samples_git_repo_url, path: _samples_dir }
          end

          def sample_exist?(sample_name)
            _sample_list.include?(sample_name)
          end

          def _sample_repo_blobs_entries(sample_name)
            _sample_git_repo
              .blobs_entries(_sample_git_repo.latest_branch_commit(_samples_git_repo_branch), paths: [_sample_directory(sample_name)])
              .reject { |_, entry| entry[:filemode] == 0o160000 }
          end

          def _sample_directory(sample_name)
            File.expand_path(File.join('/', _samples_dir, sample_name))[1..-1]
          end

          def _sample_slice_cwd(sample_name, path)
            path
              .reverse
              .chomp(_sample_directory(sample_name).reverse)
              .chomp('/')
              .reverse
          end
        end
      end
    end
  end
end
