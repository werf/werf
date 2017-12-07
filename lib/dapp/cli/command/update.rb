module Dapp
  class CLI
    module Command
      class Update < ::Dapp::CLI
        banner <<BANNER.freeze
Usage:

  dapp update

Options:
BANNER

        def run(argv = ARGV)
          self.class.parse_options(self, argv)

          spec = Gem::Specification.find do |s|
            File.fnmatch?(File.join(s.full_gem_path, '**', '*'), __FILE__, File::FNM_PATHNAME)
          end
          unless (latest_version = latest_beta_version(spec)).nil?
            try_lock do
              Gem.install(spec.name, latest_version)
            end
          end
        rescue Gem::FilePermissionError => e
          raise Errno::EACCES, e.message
        end

        def latest_beta_version(current_spec)
          minor_version = current_spec.version.approximate_recommendation
          url = "https://rubygems.org/api/v1/versions/#{current_spec.name}.json"
          response = Excon.get(url)
          if response.status == 200
            JSON.parse(response.body)
              .reduce([]) { |versions, spec| versions << Gem::Version.new(spec['number']) }
              .reject { |spec_version| minor_version != spec_version.approximate_recommendation || current_spec.version >= spec_version }
              .first
          else
            warn "Cannot get `#{url}`: got bad http status `#{response.status}`"
          end
        end

        def try_lock
          old_umask = File.umask(0)
          file = nil

          begin
            begin
              file = ::File.open('/tmp/dapp-update-running.lock', ::File::RDWR | ::File::CREAT, 0777)
            ensure
              File.umask(old_umask)
            end

            if file.flock(::File::LOCK_EX | ::File::LOCK_NB)
              yield
            else
              puts 'There are other active dapp update process, exiting without update'
            end
          ensure
            file.close unless file.nil?
          end
        end
      end
    end
  end
end
