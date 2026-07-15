require "test_helper"

class PumaHooksTest < ActiveSupport::TestCase
  class FakePumaConfig
    attr_reader :boot_callback, :shutdown_callback, :single_boot_callback, :single_shutdown_callback

    def initialize(workers: 1)
      @workers = workers
    end

    def get(key, default = nil)
      key == :workers ? @workers : default
    end

    def before_worker_boot(&block)
      @boot_callback = block
    end

    def before_worker_shutdown(&block)
      @shutdown_callback = block
    end

    def after_booted(&block)
      @single_boot_callback = block
    end

    def after_stopped(&block)
      @single_shutdown_callback = block
    end
  end

  test "registers boot and shutdown callbacks" do
    config = FakePumaConfig.new
    Observability::PumaHooks.new.register(config)
    assert config.boot_callback
    assert config.shutdown_callback
  end

  test "forwards the worker index after worker boot" do
    config = FakePumaConfig.new
    boot_indexes = []
    Observability::PumaHooks.new(worker_boot: ->(index) { boot_indexes << index }).register(config)
    config.boot_callback.call(4)
    assert_equal [4], boot_indexes
  end

  test "forwards the worker index before worker shutdown" do
    config = FakePumaConfig.new
    shutdown_indexes = []
    Observability::PumaHooks.new(worker_shutdown: ->(index) { shutdown_indexes << index }).register(config)
    config.shutdown_callback.call(5)
    assert_equal [5], shutdown_indexes
  end

  test "activates boot and shutdown in single-process mode" do
    config = FakePumaConfig.new(workers: 0)
    events = []
    Observability::PumaHooks.new(
      worker_boot: ->(index) { events << [:boot, index] },
      worker_shutdown: ->(index) { events << [:shutdown, index] }
    ).register(config)
    config.single_boot_callback.call
    config.single_shutdown_callback.call
    assert_equal [[:boot, nil], [:shutdown, nil]], events
  end

  test "puma configuration activates the runtime hooks" do
    source = File.read(Rails.root.join("config/puma.rb"))
    assert_includes source, "Observability::PumaHooks.new.register(self)"
    assert_not_includes source, "intentionally dormant"
  end
end