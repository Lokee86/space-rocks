module Observability
  class PumaHooks
    def initialize(worker_boot: ->(_worker_index) {}, worker_shutdown: ->(_worker_index) {})
      @worker_boot = worker_boot
      @worker_shutdown = worker_shutdown
    end

    def register(puma_config)
      puma_config.before_worker_boot { |worker_index| @worker_boot.call(worker_index) }
      puma_config.before_worker_shutdown { |worker_index| @worker_shutdown.call(worker_index) }
    end
  end
end