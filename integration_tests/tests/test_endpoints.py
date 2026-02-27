import pytest

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.endpoints


@pytest.fixture()
def endpoint(authenticated_client: UIAutomationClient):
    """Create a temporary endpoint for tests."""
    ep = authenticated_client.create_endpoint(
        name="Test Endpoint",
        url="https://example.com",
    )
    yield ep
    try:
        authenticated_client.delete_endpoint(ep["id"])
    except APIError:
        pass


class TestCreateEndpoint:
    def test_create_endpoint_with_defaults(
        self,
        authenticated_client: UIAutomationClient,
    ):
        resp = authenticated_client.create_endpoint(
            name="Default Endpoint",
            url="https://example.com",
        )
        assert "id" in resp
        assert resp["name"] == "Default Endpoint"
        assert resp["url"] == "https://example.com"
        # Should have default credentials
        assert isinstance(resp["credentials"], list)
        assert len(resp["credentials"]) >= 2
        # Clean up
        authenticated_client.delete_endpoint(resp["id"])

    def test_create_endpoint_with_custom_credentials(
        self,
        authenticated_client: UIAutomationClient,
    ):
        creds = [
            {"key": "api_key", "value": "test-key-123"},
            {"key": "secret", "value": "test-secret"},
        ]
        resp = authenticated_client.create_endpoint(
            name="Custom Creds Endpoint",
            url="https://api.example.com",
            credentials=creds,
        )
        assert resp["name"] == "Custom Creds Endpoint"
        assert resp["url"] == "https://api.example.com"
        assert len(resp["credentials"]) == 2
        assert resp["credentials"][0]["key"] == "api_key"
        assert resp["credentials"][0]["value"] == "test-key-123"
        # Clean up
        authenticated_client.delete_endpoint(resp["id"])

    def test_create_endpoint_missing_name(
        self,
        authenticated_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_endpoint(name="", url="https://example.com")
        assert exc_info.value.status_code == 400

    def test_create_endpoint_missing_url(
        self,
        authenticated_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_endpoint(name="No URL", url="")
        assert exc_info.value.status_code == 400

    def test_create_endpoint_unauthenticated(
        self,
        fresh_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            fresh_client.create_endpoint(
                name="Test", url="https://example.com",
            )
        assert exc_info.value.status_code == 401


class TestListEndpoints:
    def test_list_endpoints(
        self,
        authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        resp = authenticated_client.list_endpoints()
        assert "items" in resp
        assert "total" in resp
        assert resp["total"] >= 1
        ids = [e["id"] for e in resp["items"]]
        assert endpoint["id"] in ids

    def test_list_endpoints_pagination(
        self,
        authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        resp = authenticated_client.list_endpoints(limit=1, offset=0)
        assert resp["limit"] == 1
        assert resp["offset"] == 0


class TestGetEndpoint:
    def test_get_endpoint(
        self,
        authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        resp = authenticated_client.get_endpoint(endpoint["id"])
        assert resp["id"] == endpoint["id"]
        assert resp["name"] == endpoint["name"]
        assert resp["url"] == endpoint["url"]

    def test_get_endpoint_not_found(
        self,
        authenticated_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_endpoint(
                "00000000-0000-0000-0000-000000000000",
            )
        assert exc_info.value.status_code == 404


class TestUpdateEndpoint:
    def test_update_endpoint_name(
        self,
        authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        resp = authenticated_client.update_endpoint(
            endpoint["id"], name="Updated Endpoint",
        )
        assert resp["name"] == "Updated Endpoint"

    def test_update_endpoint_url(
        self,
        authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        resp = authenticated_client.update_endpoint(
            endpoint["id"], url="https://updated.example.com",
        )
        assert resp["url"] == "https://updated.example.com"

    def test_update_endpoint_credentials(
        self,
        authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        new_creds = [{"key": "token", "value": "abc123"}]
        resp = authenticated_client.update_endpoint(
            endpoint["id"], credentials=new_creds,
        )
        assert len(resp["credentials"]) == 1
        assert resp["credentials"][0]["key"] == "token"


class TestDeleteEndpoint:
    def test_delete_endpoint(
        self,
        authenticated_client: UIAutomationClient,
    ):
        ep = authenticated_client.create_endpoint(
            name="To Delete", url="https://delete.example.com",
        )
        authenticated_client.delete_endpoint(ep["id"])
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_endpoint(ep["id"])
        assert exc_info.value.status_code == 404


class TestEndpointOwnership:
    def test_other_user_cannot_access_endpoint(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        with pytest.raises(APIError) as exc_info:
            second_authenticated_client.get_endpoint(endpoint["id"])
        assert exc_info.value.status_code == 403

    def test_other_user_cannot_update_endpoint(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        with pytest.raises(APIError) as exc_info:
            second_authenticated_client.update_endpoint(
                endpoint["id"], name="Hacked",
            )
        assert exc_info.value.status_code == 403

    def test_other_user_cannot_delete_endpoint(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
        endpoint: dict,
    ):
        with pytest.raises(APIError) as exc_info:
            second_authenticated_client.delete_endpoint(endpoint["id"])
        assert exc_info.value.status_code == 403
