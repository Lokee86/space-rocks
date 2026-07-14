module Observability
  ProcessIdentity = Data.define(:service_instance_id, :worker_id, :pid) do
    def self.resolve(service_instance_id:, worker_id: nil, env: ENV, pid: Process.pid)
      resolved_worker_id = worker_id || env["PUMA_WORKER_INDEX"] || env["API_WORKER_ID"]
      resolved_worker_id = "single-process" if resolved_worker_id.nil? || resolved_worker_id.to_s.empty?
      resolved_worker_id = "worker-#{resolved_worker_id}" if worker_id.nil? && resolved_worker_id != "single-process"

      new(service_instance_id: service_instance_id, worker_id: resolved_worker_id.to_s, pid: pid)
    end
  end
end