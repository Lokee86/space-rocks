require "thread"
require_relative "worker_runtime"

module Observability
  class PumaHooks
    @runtime_mutex = Mutex.new

    class << self
      def boot(worker_index)
        runtime = @runtime_mutex.synchronize do
          reset_inherited_runtime
          unless @runtime
            @runtime = WorkerRuntime.new
            @runtime_pid = Process.pid
          end
          @runtime
        end
        runtime.boot(worker_index: worker_index)
      end

      def shutdown(_worker_index)
        runtime = @runtime_mutex.synchronize do
          reset_inherited_runtime
          active_runtime = @runtime
          @runtime = nil
          @runtime_pid = nil
          active_runtime
        end
        runtime&.shutdown
      end

      def status
        @runtime_mutex.synchronize do
          @runtime&.status || { enabled: false, booted: false, shutdown: true, degraded: false }
        end
      end

      private

      def reset_inherited_runtime
        return if @runtime_pid.nil? || @runtime_pid == Process.pid

        @runtime = nil
        @runtime_pid = nil
      end
    end

    def initialize(worker_boot: ->(worker_index) { self.class.boot(worker_index) }, worker_shutdown: ->(worker_index) { self.class.shutdown(worker_index) })
      @worker_boot = worker_boot
      @worker_shutdown = worker_shutdown
    end

    def register(puma_config)
      puma_config.before_worker_boot { |worker_index| @worker_boot.call(worker_index) }
      puma_config.before_worker_shutdown { |worker_index| @worker_shutdown.call(worker_index) }
      register_single_process_hooks(puma_config) if single_process?(puma_config)
    end

    private

    def single_process?(puma_config)
      puma_config.respond_to?(:get) && puma_config.get(:workers, 0).to_i.zero?
    end

    def register_single_process_hooks(puma_config)
      puma_config.after_booted { @worker_boot.call(nil) }
      puma_config.after_stopped { @worker_shutdown.call(nil) }
    end
  end
end