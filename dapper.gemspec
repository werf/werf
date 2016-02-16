Gem::Specification.new do |s|
  s.name = 'dapper'
  s.version = '0.0.1'
  s.date = '2016-01-22'

  s.summary = ''
  s.description = s.summary
  s.homepage = 'https://github.com/flant/dapper'

  s.authors = ['Dmitry Stolyarov']
  s.email = ''
  s.license = ''

  s.files = Dir['lib/**/*']
  s.executables = ['dappit']

  s.add_dependency 'mixlib-shellout', '>= 1.0', '< 3.0'
  s.add_dependency 'docopt', '>= 0.5', '< 2.0'
end
