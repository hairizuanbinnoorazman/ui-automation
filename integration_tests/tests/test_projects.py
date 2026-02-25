import pytest

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.projects


@pytest.fixture()
def project(authenticated_client: UIAutomationClient):
    """Create a temporary project and delete it after the test."""
    p = authenticated_client.create_project(
        name="Test Project",
        description="Created by pytest fixture",
    )
    yield p
    try:
        authenticated_client.delete_project(p["id"])
    except APIError:
        pass


class TestCreateProject:
    def test_create_project(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_project(
            name="Integration Project",
            description="Created by integration tests",
        )
        assert "id" in resp
        assert resp["name"] == "Integration Project"
        assert resp["description"] == "Created by integration tests"
        # cleanup
        authenticated_client.delete_project(resp["id"])


class TestListProjects:
    def test_list_projects_shape(
        self,
        authenticated_client: UIAutomationClient,
        project: dict,
    ):
        resp = authenticated_client.list_projects(limit=10, offset=0)
        assert "items" in resp
        assert "total" in resp
        assert "limit" in resp
        assert "offset" in resp
        assert isinstance(resp["items"], list)
        assert resp["total"] >= 1


class TestGetProject:
    def test_get_project_by_id(
        self,
        authenticated_client: UIAutomationClient,
        project: dict,
    ):
        resp = authenticated_client.get_project(project["id"])
        assert resp["id"] == project["id"]
        assert resp["name"] == project["name"]

    def test_get_nonexistent_project(
        self, authenticated_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_project(
                "00000000-0000-0000-0000-000000000000",
            )
        assert exc_info.value.status_code in (403, 404)


class TestUpdateProject:
    def test_update_project_name(
        self,
        authenticated_client: UIAutomationClient,
        project: dict,
    ):
        resp = authenticated_client.update_project(
            project["id"], name="Updated Name",
        )
        assert resp["name"] == "Updated Name"


class TestDeleteProject:
    def test_delete_project(self, authenticated_client: UIAutomationClient):
        p = authenticated_client.create_project(
            name="To Delete", description="Will be deleted",
        )
        resp = authenticated_client.delete_project(p["id"])
        assert "message" in resp

        # Verify it's gone
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_project(p["id"])
        assert exc_info.value.status_code in (403, 404)
