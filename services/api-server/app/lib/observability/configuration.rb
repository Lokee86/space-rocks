module Observability
  Configuration = Data.define(
    :enabled,
    :log_root,
    :service_instance_id,
    :segment_bytes,
    :segment_age,
    :retention_age,
    :retention_bytes,
    :compression
  )
end