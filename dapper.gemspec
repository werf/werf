Gem::Specification.new do |s|
  s.name = 'dapper'
  s.version = Dapper.VERSION

  s.summary = 'Build docker packaged apps using chef or shell'
  s.description = s.summary
  s.homepage = 'https://github.com/flant/dapper'

  s.authors = ['Dmitry Stolyarov']
  s.email = 'dmitry.stolyarov@flant.com'
  s.license = 'MIT'

  s.files = Dir['lib/**/*']
  s.executables = ['dappit']

  s.add_dependency 'mixlib-shellout', '>= 1.0', '< 3.0'
  s.add_dependency 'docopt', '>= 0.5', '< 2.0'
end
