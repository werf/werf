module Dapp
  module Downloader
    module Error
      class DownloadFailed < ::Exception
      end
    end

    class BytesCount
      attr_reader :bytes, :total_bytes_count

      def initialize(bytes, total_bytes_count: nil)
        @bytes = bytes.to_f
        @total_bytes_count = total_bytes_count
      end

      def to_s(*a)
        max_bytes = @total_bytes_count || self
        width = sprintf("%.2f", max_bytes.bytes/1024/1024).bytesize
        sprintf("%#{width}.2f", @bytes/1024/1024)
      end

      [:+, :-, :*, :/].each do |method|
        define_method(method) do |arg|
          res = case arg
          when BytesCount
            @bytes.send(method, arg.bytes)
          else
            @bytes.send(method, arg)
          end
          self.class.new(res, total_bytes_count: total_bytes_count)
        end
      end

      [:>, :<, :==].each do |method|
        define_method(method) do |arg|
          case arg
          when BytesCount
            @bytes.send(method, arg.bytes)
          else
            @bytes.send(method, arg)
          end
        end
      end

      def method_missing(method, *args, &blk)
        case method
        when :to_f, :to_i
          @bytes.send(method, *args, &blk)
        else
          raise
        end
      end
    end # BytesCount

    class << self
      def download(url, destination, show_progress: false, progress_titile: nil)
        resp = nil
        location = URI(url)
        done = false
        state = {}

        loop do
          Net::HTTP.start(location.host, location.port, use_ssl: true) do |http|
            req = Net::HTTP::Get.new location
            http.request req do |resp|
              case resp
              when Net::HTTPRedirection
                location = URI(resp["location"])
                next
              when Net::HTTPSuccess
                File.open(destination, "wb") do |file|
                  file_size_mb = nil
                  file_size_mb = BytesCount.new(resp.to_hash["content-length"].first.to_f) if show_progress

                  if show_progress
                    old_DEFAULT_BEGINNING_POSITION = ProgressBar::Progress.send(:remove_const, :DEFAULT_BEGINNING_POSITION)
                    ProgressBar::Progress.send(:const_set, :DEFAULT_BEGINNING_POSITION, BytesCount.new(0, total_bytes_count: file_size_mb))
                  end

                  begin
                    progressbar = nil
                    progressbar = ProgressBar.create(format: "   %cMB / %CMB   %B  %t", starting_at: BytesCount.new(0, total_bytes_count: file_size_mb), total: file_size_mb, progress_mark: "#", remainder_mark: ".", title: progress_titile, length: 100, autofinish: true) if show_progress

                    resp.read_body do |segment|
                      progressbar.progress = progressbar.progress + segment.bytesize if show_progress
                      file.write segment
                    end
                  ensure
                    if show_progress
                      ProgressBar::Progress.send(:remove_const, :DEFAULT_BEGINNING_POSITION)
                      ProgressBar::Progress.send(:const_set, :DEFAULT_BEGINNING_POSITION, old_DEFAULT_BEGINNING_POSITION)
                    end
                  end
                end # File.open

                done = true
              else
                raise Error::DownloadFailed, "Failed to download #{url}: #{resp.code} #{resp.message}"
              end # when
            end # http.request
          end # Net::HTTP.start

          break if done
        end # loop
      end
    end
  end # Downloader
end # Dapp
