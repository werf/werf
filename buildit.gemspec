Gem::Specification.new do |s|
  s.name = 'buildit'
  s.version = '0.0.0'
  s.date = '2016-01-22'

  s.summary = ''
  s.description = s.summary
  s.homepage = 'https://github.com/flant/buildit'

  s.authors = ['']
  s.email = ''
  s.license = ''

  s.files = Dir['lib/**/*']
  s.executables = ['buildit']
  s.add_dependency 'mixlib-shellout', '>= 1.0', '< 3.0'
end
