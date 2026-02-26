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


class TestAssignUser:
    def test_default_assigned_to_is_null(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        fetched = authenticated_client.get_run(run["id"])
        assert fetched.get("assigned_to") is None

    def test_assign_user_to_run(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        me = authenticated_client.me()
        updated = authenticated_client.assign_run(run["id"], me["id"])
        assert updated["assigned_to"] == me["id"]

    def test_unassign_user_from_run(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        me = authenticated_client.me()
        authenticated_client.assign_run(run["id"], me["id"])
        updated = authenticated_client.unassign_run(run["id"])
        assert updated.get("assigned_to") is None

    def test_assign_invalid_user_returns_error(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        with pytest.raises(APIError) as exc_info:
            authenticated_client.assign_run(
                run["id"], "00000000-0000-0000-0000-000000000000",
            )
        assert exc_info.value.status_code == 400

    def test_assign_malformed_uuid_returns_400(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        with pytest.raises(APIError) as exc_info:
            authenticated_client.assign_run(run["id"], "not-a-uuid")
        assert exc_info.value.status_code == 400

    def test_reassign_user(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        me = authenticated_client.me()
        second_user = second_authenticated_client.me()
        authenticated_client.assign_run(run["id"], me["id"])
        updated = authenticated_client.assign_run(run["id"], second_user["id"])
        assert updated["assigned_to"] == second_user["id"]

    def test_assign_persists_on_get(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        me = authenticated_client.me()
        authenticated_client.assign_run(run["id"], me["id"])
        fetched = authenticated_client.get_run(run["id"])
        assert fetched["assigned_to"] == me["id"]

    def test_assigned_to_survives_notes_update(
        self,
        authenticated_client: UIAutomationClient,
        project_and_procedure: tuple,
    ):
        _, procedure = project_and_procedure
        run = authenticated_client.create_run(procedure["id"])
        me = authenticated_client.me()
        authenticated_client.assign_run(run["id"], me["id"])
        updated = authenticated_client.update_run(run["id"], notes="some notes")
        assert updated["assigned_to"] == me["id"]
        assert updated["notes"] == "some notes"


class TestAssignUserAuthorization:
    @pytest.fixture()
    def other_user_project_and_run(
        self,
        authenticated_client: UIAutomationClient,
    ):
        """Create a project, procedure, and run owned by the primary user."""
        project = authenticated_client.create_project(
            name="Auth Test Project",
            description="For authorization tests",
        )
        procedure = authenticated_client.create_procedure(
            project_id=project["id"],
            name="Auth Test Procedure",
            description="Procedure for auth tests",
            steps=SAMPLE_STEPS,
        )
        run = authenticated_client.create_run(procedure["id"])
        yield project, run
        try:
            authenticated_client.delete_project(project["id"])
        except APIError:
            pass

    def test_other_user_cannot_assign_run(
        self,
        second_authenticated_client: UIAutomationClient,
        other_user_project_and_run: tuple,
    ):
        _, run = other_user_project_and_run
        second_user = second_authenticated_client.me()
        with pytest.raises(APIError) as exc_info:
            second_authenticated_client.assign_run(run["id"], second_user["id"])
        assert exc_info.value.status_code == 403

    def test_other_user_cannot_unassign_run(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
        other_user_project_and_run: tuple,
    ):
        _, run = other_user_project_and_run
        me = authenticated_client.me()
        authenticated_client.assign_run(run["id"], me["id"])
        with pytest.raises(APIError) as exc_info:
            second_authenticated_client.unassign_run(run["id"])
        assert exc_info.value.status_code == 403


class TestUserSearch:
    def test_search_users(
        self,
        authenticated_client: UIAutomationClient,
    ):
        me = authenticated_client.me()
        result = authenticated_client.list_users(search=me["username"])
        assert "users" in result
        assert len(result["users"]) >= 1
        usernames = [u["username"] for u in result["users"]]
        assert me["username"] in usernames

    def test_search_users_no_results(
        self,
        authenticated_client: UIAutomationClient,
    ):
        result = authenticated_client.list_users(
            search="nonexistent_user_xyz_12345",
        )
        assert "users" in result
        assert len(result["users"]) == 0

    def test_list_users_without_search_param(
        self,
        authenticated_client: UIAutomationClient,
    ):
        result = authenticated_client.list_users()
        assert "users" in result
        assert len(result["users"]) >= 1
