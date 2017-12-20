module Dapp
  module Deployment
    class Secret
      attr_reader :key

      def initialize(key)
        self.class._validate_key!(key)
        @key = key
      end

      def generate(value)
        cipher = self.class._openssl_cipher
        cipher.encrypt
        cipher.key = self.class._hex_to_binary key
        iv = cipher.random_iv

        iv_size_prefix = [iv.bytesize].pack('S')
        encrypted = cipher.update(value.to_s) + cipher.final

        self.class._binary_to_hex "#{iv_size_prefix}#{iv}#{encrypted}"
      end

      def extract(hexdata)
        data = self.class._hex_to_binary hexdata.to_s

        iv_size = data.unpack('S').first
        data = data.byteslice(2..-1)
        raise ExtractionError, code: :bad_data, data: {data: hexdata} unless data

        iv = data.byteslice(0, iv_size)
        data = data.byteslice(iv_size..-1)
        raise ExtractionError, code: :bad_data, data: {data: hexdata} unless data

        decipher = self.class._openssl_cipher
        decipher.decrypt
        decipher.key = self.class._hex_to_binary(key)

        begin
          decipher.iv = iv
        rescue OpenSSL::Cipher::CipherError
          raise ExtractionError, code: :bad_data, data: {data: hexdata}
        end

        begin
          value = decipher.update(data) + decipher.final
        rescue OpenSSL::Cipher::CipherError
          raise ExtractionError, code: :bad_data, data: {data: hexdata}
        end
        value.force_encoding('utf-8')
      end

      class << self
        def generate_key
          _binary_to_hex _openssl_cipher.random_key
        end

        def _openssl_cipher
          OpenSSL::Cipher::AES.new(128, :CBC)
        end

        def _hex_to_binary(key)
          [key].pack('H*')
        end

        def _binary_to_hex(key)
          key.unpack('H*').first
        end

        def _validate_key!(key)
          # Требуется 128 битный ключ — это 16 байт.
          # Ключ закодирован в hex кодировке для пользователя.
          # 2 hex символа на 1 байт в hex кодировке.
          # Поэтому требуется длина ключа в hex кодировке в 32 символа.
          if key.bytesize < 32
            raise InvalidKeyError, code: :key_length_too_short, data: {required_size: 32}
          end
        end
      end

      class Error < ::Dapp::Deployment::Error::Default
        def initialize(**net_status)
          super(net_status.merge(context: :secret))
        end
      end

      class InvalidKeyError < Error
      end

      class ExtractionError < Error
      end
    end
  end
end
