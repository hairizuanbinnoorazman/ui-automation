import pytest
import uuid

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.integrations


@pytest.fixture(scope="module")
def integration_project(authenticated_client: UIAutomationClient) -> dict:
    """Create a test project for integration tests."""
    return authenticated_client.create_project(
        name=f"integration-test-{uuid.uuid4().hex[:8]}",
        description="Project for integration tests",
    )


@pytest.fixture(scope="module")
def integration_procedure(
    authenticated_client: UIAutomationClient,
    integration_project: dict,
) -> dict:
    """Create a test procedure."""
    return authenticated_client.create_procedure(
        project_id=integration_project["id"],
        name="Integration Test Procedure",
        steps=[{"name": "Step 1", "instructions": "Do something", "image_paths": []}],
    )


@pytest.fixture(scope="module")
def integration_run(
    authenticated_client: UIAutomationClient,
    integration_procedure: dict,
) -> dict:
    """Create a test run."""
    return authenticated_client.create_run(
        procedure_id=integration_procedure["id"],
    )


class TestCreateIntegration:
    def test_create_github_integration(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_integration(
            name="Test GitHub",
            provider="github",
            credentials=[
                {"key": "token", "value": "ghp_fake_token_12345"},
            ],
        )
        assert "id" in resp
        assert resp["name"] == "Test GitHub"
        assert resp["provider"] == "github"
        assert resp["is_active"] is True
        # Cleanup
        authenticated_client.delete_integration(resp["id"])

    def test_create_jira_integration(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_integration(
            name="Test Jira",
            provider="jira",
            credentials=[
                {"key": "url", "value": "https://example.atlassian.net"},
                {"key": "email", "value": "test@example.com"},
                {"key": "api_token", "value": "fake_token"},
            ],
        )
        assert "id" in resp
        assert resp["name"] == "Test Jira"
        assert resp["provider"] == "jira"
        authenticated_client.delete_integration(resp["id"])

    def test_create_invalid_provider(self, authenticated_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_integration(
                name="Invalid",
                provider="invalid_provider",
                credentials=[],
            )
        assert exc_info.value.status_code == 400

    def test_create_missing_name(self, authenticated_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_integration(
                name="",
                provider="github",
                credentials=[{"key": "token", "value": "test"}],
            )
        assert exc_info.value.status_code == 400

    def test_create_unauthenticated(self, fresh_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            fresh_client.create_integration(
                name="Test",
                provider="github",
                credentials=[{"key": "token", "value": "test"}],
            )
        assert exc_info.value.status_code == 401


class TestListIntegrations:
    def test_list_empty(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.list_integrations()
        assert "items" in resp
        assert "total" in resp
        assert isinstance(resp["items"], list)

    def test_list_after_create(self, authenticated_client: UIAutomationClient):
        # Create one
        created = authenticated_client.create_integration(
            name="List Test",
            provider="github",
            credentials=[{"key": "token", "value": "test"}],
        )
        resp = authenticated_client.list_integrations()
        ids = [i["id"] for i in resp["items"]]
        assert created["id"] in ids
        # Cleanup
        authenticated_client.delete_integration(created["id"])


class TestGetIntegration:
    def test_get_success(self, authenticated_client: UIAutomationClient):
        created = authenticated_client.create_integration(
            name="Get Test",
            provider="github",
            credentials=[{"key": "token", "value": "test"}],
        )
        resp = authenticated_client.get_integration(created["id"])
        assert resp["id"] == created["id"]
        assert resp["name"] == "Get Test"
        authenticated_client.delete_integration(created["id"])

    def test_get_not_found(self, authenticated_client: UIAutomationClient):
        fake_id = str(uuid.uuid4())
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_integration(fake_id)
        assert exc_info.value.status_code == 404


class TestDeleteIntegration:
    def test_delete_success(self, authenticated_client: UIAutomationClient):
        created = authenticated_client.create_integration(
            name="Delete Test",
            provider="github",
            credentials=[{"key": "token", "value": "test"}],
        )
        authenticated_client.delete_integration(created["id"])
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_integration(created["id"])
        assert exc_info.value.status_code == 404


class TestIssueLinks:
    """Test issue link operations. Note: creating/resolving actual external issues
    requires real provider credentials, so we test the link/unlink flow using
    the link-existing endpoint which doesn't call external APIs."""

    def test_list_empty(
        self,
        authenticated_client: UIAutomationClient,
        integration_run: dict,
    ):
        resp = authenticated_client.list_issue_links(integration_run["id"])
        assert isinstance(resp, list)
        assert len(resp) == 0

    def test_link_existing_fails_with_fake_credentials(
        self,
        authenticated_client: UIAutomationClient,
        integration_run: dict,
    ):
        """link_existing_issue calls the external API to validate the issue.
        With fake credentials, the external API rejects the request."""
        integ = authenticated_client.create_integration(
            name="Link Test Integration",
            provider="github",
            credentials=[{"key": "token", "value": "ghp_fake_token_12345"}],
        )

        with pytest.raises(APIError) as exc_info:
            authenticated_client.link_existing_issue(
                run_id=integration_run["id"],
                integration_id=integ["id"],
                external_id="owner/repo#42",
            )
        # External API call fails => 500 from our backend
        assert exc_info.value.status_code == 500

        # Cleanup
        authenticated_client.delete_integration(integ["id"])

    def test_unlink_not_found(
        self,
        authenticated_client: UIAutomationClient,
        integration_run: dict,
    ):
        """Unlinking a non-existent issue link returns 404."""
        fake_link_id = str(uuid.uuid4())
        with pytest.raises(APIError) as exc_info:
            authenticated_client.unlink_issue(integration_run["id"], fake_link_id)
        assert exc_info.value.status_code == 404

    def test_link_unauthenticated(
        self,
        fresh_client: UIAutomationClient,
        integration_run: dict,
    ):
        with pytest.raises(APIError) as exc_info:
            fresh_client.list_issue_links(integration_run["id"])
        assert exc_info.value.status_code == 401
