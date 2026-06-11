require "openapi_first"

module OpenapiContractAssertions
  def openapi_definition
    @openapi_definition ||= OpenapiFirst.load(Rails.root.join("../../shared/contracts/http/openapi.yaml").expand_path)
  end

  def assert_openapi_request!
    validated_request = openapi_definition.validate_request(openapi_request, raise_error: false)
    assert validated_request.valid?, format_openapi_failure(validated_request)
  end

  def assert_openapi_response!
    validated_response = openapi_definition.validate_response(openapi_request, openapi_response, raise_error: false)
    assert validated_response.valid?, format_openapi_failure(validated_response)
  end

  def assert_openapi_contract!
    assert_openapi_request!
    assert_openapi_response!
  end

  private

  def openapi_request
    Rack::Request.new(@request.env)
  end

  def openapi_response
    Rack::Response.new(@response.body, @response.status, @response.headers.to_h)
  end

  def format_openapi_failure(validated)
    validated.error&.exception_message || validated.error&.exception(validated)&.message || "OpenAPI contract validation failed"
  end
end
