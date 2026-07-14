require "test_helper"

class ProcessIdentityTest < ActiveSupport::TestCase
  Identity = Observability::ProcessIdentity

  test "explicit worker identity wins" do
    identity = Identity.resolve(service_instance_id: "api-a", worker_id: "blue", env: { "PUMA_WORKER_INDEX" => "2" }, pid: 12)

    assert_equal ["api-a", "blue", 12], identity.to_h.values
  end

  test "uses Puma worker index from the environment" do
    identity = Identity.resolve(service_instance_id: "api-a", env: { "PUMA_WORKER_INDEX" => "2" }, pid: 12)

    assert_equal "worker-2", identity.worker_id
  end

  test "uses a single-process identity without a worker index" do
    identity = Identity.resolve(service_instance_id: "api-a", env: {}, pid: 12)

    assert_equal "single-process", identity.worker_id
  end
end