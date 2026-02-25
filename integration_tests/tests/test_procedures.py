import pytest

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.procedures

SAMPLE_STEPS = [
    {
        "name": "Open login page",
        "instructions": "Navigate to https://example.com/login",
        "image_paths": [],
    },
    {
        "name": "Enter credentials",
        "instructions": "Type username and password into the form fields",
        "image_paths": [],
    },
    {
        "name": "Submit form",
        "instructions": "Click the login button and verify redirect",
        "image_paths": [],
    },
]


@pytest.fixture()
def project_id(authenticated_client: UIAutomationClient):
    """Create a temporary project for procedure tests."""
    p = authenticated_client.create_project(
        name="Procedure Test Project",
        description="For procedure integration tests",
    )
    yield p["id"]
    try:
        authenticated_client.delete_project(p["id"])
    except APIError:
        pass


@pytest.fixture()
def procedure(authenticated_client: UIAutomationClient, project_id: str):
    """Create a temporary test procedure."""
    return authenticated_client.create_procedure(
        project_id=project_id,
        name="Login Test Procedure",
        description="Test login functionality",
        steps=SAMPLE_STEPS,
    )


class TestCreateProcedure:
    def test_create_procedure_with_steps(
        self,
        authenticated_client: UIAutomationClient,
        project_id: str,
    ):
        resp = authenticated_client.create_procedure(
            project_id=project_id,
            name="New Procedure",
            description="A test procedure",
            steps=SAMPLE_STEPS,
        )
        assert "id" in resp
        assert resp["name"] == "New Procedure"
        assert resp["description"] == "A test procedure"
        assert resp["version"] == 1
        assert resp["is_latest"] is True


class TestListProcedures:
    def test_list_procedures(
        self,
        authenticated_client: UIAutomationClient,
        project_id: str,
        procedure: dict,
    ):
        resp = authenticated_client.list_procedures(project_id)
        assert "items" in resp
        assert "total" in resp
        assert resp["total"] >= 1
        ids = [p["id"] for p in resp["items"]]
        assert procedure["id"] in ids


class TestGetProcedure:
    def test_get_procedure(
        self,
        authenticated_client: UIAutomationClient,
        project_id: str,
        procedure: dict,
    ):
        resp = authenticated_client.get_procedure(
            project_id, procedure["id"],
        )
        assert resp["id"] == procedure["id"]
        assert resp["name"] == procedure["name"]


class TestUpdateProcedure:
    def test_update_in_place(
        self,
        authenticated_client: UIAutomationClient,
        project_id: str,
        procedure: dict,
    ):
        resp = authenticated_client.update_procedure(
            project_id,
            procedure["id"],
            description="Updated description",
        )
        assert resp["description"] == "Updated description"
        # Update endpoint now returns the draft (version 0)
        assert resp["version"] == 0


class TestVersioning:
    def test_create_version(
        self,
        authenticated_client: UIAutomationClient,
        project_id: str,
        procedure: dict,
    ):
        new_version = authenticated_client.create_version(
            project_id, procedure["id"],
        )
        assert "id" in new_version
        assert new_version["id"] != procedure["id"]
        assert new_version["version"] == procedure["version"] + 1
        assert new_version["is_latest"] is True

    def test_version_history(
        self,
        authenticated_client: UIAutomationClient,
        project_id: str,
    ):
        # Create a fresh procedure so we control the full history
        proc = authenticated_client.create_procedure(
            project_id=project_id,
            name="Versioned Procedure",
            description="For version history test",
            steps=SAMPLE_STEPS,
        )
        # Create a second version
        authenticated_client.create_version(project_id, proc["id"])

        history = authenticated_client.get_version_history(
            project_id, proc["id"],
        )
        assert isinstance(history, list)
        assert len(history) >= 2
        versions = [h["version"] for h in history]
        assert 1 in versions
        assert 2 in versions
