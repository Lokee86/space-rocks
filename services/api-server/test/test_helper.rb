ENV["RAILS_ENV"] ||= "test"
require_relative "../config/environment"
require "rails/test_help"
require_relative "support/openapi_contract_assertions"

ActionDispatch::IntegrationTest.include(OpenapiContractAssertions)

module ActiveSupport
  class TestCase
    # Run tests in parallel with specified workers
    parallelize(workers: :number_of_processors)

    # Setup all fixtures in test/fixtures/*.yml for all tests in alphabetical order.
    fixtures :all

    def with_singleton_method_stub(target, method_name, replacement)
      singleton_class = target.singleton_class
      original_method = target.method(method_name)
      visibility =
        if singleton_class.private_method_defined?(method_name)
          :private
        elsif singleton_class.protected_method_defined?(method_name)
          :protected
        else
          :public
        end

      singleton_class.define_method(method_name, &replacement)
      singleton_class.send(visibility, method_name)

      yield
    ensure
      singleton_class.define_method(method_name, original_method)
      singleton_class.send(visibility, method_name)
    end
  end
end
