lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'dapp/version'

Gem::Specification.new do |s|
  s.name = 'dapp'
  s.version = Dapp::VERSION

  s.summary = 'Build docker packaged apps using chef or shell'
  s.description = s.summary
  s.homepage = 'https://github.com/flant/dapp'

  s.authors = ['Dmitry Stolyarov']
  s.email = ['dmitry.stolyarov@flant.com']
  s.license = 'MIT'

  s.files = Dir['lib/**/*', 'config/**/*']
  s.executables = ['dapp']

  s.required_ruby_version = '>= 2.1'
  s.required_rubygems_version = '>= 2.5.0'

  s.add_dependency 'mixlib-shellout', '>= 1.0', '< 3.0'
  s.add_dependency 'mixlib-cli', '>= 1.0', '< 3.0'
  s.add_dependency 'excon', '>= 0.45.4', '< 1.0'
  s.add_dependency 'net_status', '>= 0.1.2', '< 1.0'
  s.add_dependency 'i18n', '~> 0.7'
  s.add_dependency 'paint', '~> 1.0', '>= 1.0.1'
  s.add_dependency 'inifile', '~> 3.0.0'
  s.add_dependency 'rugged', '~> 0.24.0'
  s.add_dependency 'murmurhash3', '~> 0.1.6'

  s.add_development_dependency 'bundler', '~> 1.7'
  s.add_development_dependency 'rake', '~> 10.0'
  s.add_development_dependency 'rspec', '~> 3.4', '>= 3.4.0'
  s.add_development_dependency 'test_construct', '~> 2'
  s.add_development_dependency 'timecop', '~> 0.8'
  s.add_development_dependency 'pry', '>= 0.10.3', '< 1.0'
  s.add_development_dependency 'pry-stack_explorer', '>= 0.4.9.2', '< 1.0'
  s.add_development_dependency 'travis', '~> 1.8', '>= 1.8.2'
  s.add_development_dependency 'codeclimate-test-reporter', '~> 0.5'
  s.add_development_dependency 'activesupport', '~> 4.2', '>= 4.2.6'
  s.add_development_dependency 'recursive-open-struct', '~> 1.0', '>= 1.0.1'
  s.add_development_dependency 'ruby-prof', '>= 0.15.9', '< 1.0'
  s.add_development_dependency 'rubocop', '~> 0.47.0'
end
