import pytest

from client import (
    APIError,
    STATUS_FAILED,
    STATUS_PASSED,
    STATUS_PENDING,
    STATUS_RUNNING,
    UIAutomationClient,
)

pytestmark = pytest.mark.runs

SAMPLE_STEPS = [
    {
        "name": "Step 1",
        "instructions": "Do something",
        "image_paths": [],
    },
]


@pytest.fixture()
def project_and_procedure(authenticated_client: UIAutomationClient):
    """Create a project + procedure for test run tests."""
    project = authenticated_client.create_project(
        name="Run Test Project",
        description="For test run integration tests",
    )
    procedure = authenticated_client.create_procedure(
        project_id=project["id"],
        name="Run Test Procedure",
        description="Procedure for run tests",
        steps=SAMPLE_STEPS,
    )
    yield project, procedure
    try:
        authenticated_client.delete_project(project["id"])
    except APIError:
        pass


class TestCreateRun:
    def test_create_run(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        assert "id" in run
        assert run["status"] == STATUS_PENDING
        assert run["test_procedure_id"] is not None


class TestRunLifecycle:
    def test_start_run(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        started = authenticated_client.start_run(run["id"])
        assert started["status"] == STATUS_RUNNING
        assert started["started_at"] is not None

    def test_complete_run_passed(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        authenticated_client.start_run(run["id"])
        completed = authenticated_client.complete_run(
            run["id"], status=STATUS_PASSED, notes="All tests passed",
        )
        assert completed["status"] == STATUS_PASSED
        assert completed["completed_at"] is not None
        assert completed["notes"] == "All tests passed"

    def test_complete_run_failed(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        authenticated_client.start_run(run["id"])
        completed = authenticated_client.complete_run(
            run["id"], status=STATUS_FAILED, notes="Step 2 failed",
        )
        assert completed["status"] == STATUS_FAILED


class TestListRuns:
    def test_list_runs(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        authenticated_client.create_run(procedure["id"])
        resp = authenticated_client.list_runs(procedure["id"])
        assert "items" in resp
        assert "total" in resp
        assert resp["total"] >= 1


class TestGetRun:
    def test_get_run_by_id(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        fetched = authenticated_client.get_run(run["id"])
        assert fetched["id"] == run["id"]
        assert fetched["status"] == STATUS_PENDING


class TestUpdateRun:
    def test_update_run_notes(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        updated = authenticated_client.update_run(
            run["id"], notes="Updated notes",
        )
        assert updated["notes"] == "Updated notes"
