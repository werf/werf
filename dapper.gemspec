lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'dapper/version'

Gem::Specification.new do |s|
  s.name = 'dapper'
  s.version = Dapper::VERSION

  s.summary = 'Build docker packaged apps using chef or shell'
  s.description = s.summary
  s.homepage = 'https://github.com/flant/dapper'

  s.authors = ['Dmitry Stolyarov', 'Timofey Kirillov']
  s.email = ['dmitry.stolyarov@flant.com', 'timofey.kirillov@flant.com']
  s.license = 'MIT'

  s.files = Dir['lib/**/*']
  s.executables = ['dappit']

  s.required_ruby_version = '>= 2.2'

  s.add_dependency 'mixlib-shellout', '>= 1.0', '< 3.0'
  s.add_dependency 'mixlib-cli', '>= 1.0', '< 3.0'

  s.add_development_dependency 'bundler', '~> 1.7'
  s.add_development_dependency 'rake', '~> 10.0'
  s.add_development_dependency 'rspec', '~> 3.4', '>= 3.4.0'
  s.add_development_dependency 'pry', '>= 0.10.3', '< 1.0'
  s.add_development_dependency 'pry-stack_explorer', '>= 0.4.9.2', '< 1.0'
  s.add_development_dependency 'travis', '~> 1.8', '>= 1.8.2'
end
