require 'bundler/gem_tasks'
require 'rspec'
require 'rspec/core/rake_task'

def _set_process_name(name)
  $0 = name.to_s

  require "fiddle"
  RUBY_PLATFORM.index("linux") or return
  Fiddle::Function.new(
    Fiddle::Handle["prctl"], [
      Fiddle::TYPE_INT, Fiddle::TYPE_VOIDP,
      Fiddle::TYPE_LONG, Fiddle::TYPE_LONG,
      Fiddle::TYPE_LONG
    ], Fiddle::TYPE_INT
  ).call(15, name.to_s, 0, 0, 0)
end

RSpec::Core::RakeTask
  .new(:spec) do |t|
    t.pattern = ENV['TEST_PATTERN'] if ENV['TEST_PATTERN']
    t.exclude_pattern = ENV['TEST_PATTERN_EXCLUDE'] if ENV['TEST_PATTERN_EXCLUDE']

    t.instance_variable_set(:@_skip_task, true) if ENV['TRAVIS_COMMIT_MESSAGE'] =~ /\[ci skip tests?\]/
  end
  .define_singleton_method(:run_task) do |*args|
    if @_skip_task
      puts "Rake :spec task skipped"
    else
      RSpec::Core::RakeTask.instance_method(:run_task).bind(self).call(*args)
    end
  end

task :parallel_spec do
  workers = 0
  status = 0

  [[nil, 'spec/integration/chef_spec.rb,spec/integration/dimg_spec.rb,spec/integration/dimg_dev_mod_spec.rb'],
   ['spec/integration/chef_spec.rb', nil],
   ['spec/integration/dimg_spec.rb', nil],
   ['spec/integration/dimg_dev_mod_spec.rb', nil],
  ].each do |pattern, exclude_pattern|
    pid = Process.fork

    if pid
      workers += 1
    else
      _set_process_name "parallel_spec [pattern=#{pattern},exclude_pattern=#{exclude_pattern}]"

      ENV['TEST_PATTERN'] = pattern
      ENV['TEST_PATTERN_EXCLUDE'] = exclude_pattern

      Rake::Task['spec'].invoke

      exit 0
    end
  end

  while workers > 0
    status += Process.wait2[1].exitstatus
    workers -= 1
  end

  exit status if status != 0
end

task :skip do
end

task default: :skip
