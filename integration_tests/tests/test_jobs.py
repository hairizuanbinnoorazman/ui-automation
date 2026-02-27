import time

import pytest

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.jobs


@pytest.fixture()
def project_for_jobs(authenticated_client: UIAutomationClient):
    """Create a temporary project for job tests."""
    p = authenticated_client.create_project(
        name="Job Test Project",
        description="For job integration tests",
    )
    yield p
    try:
        authenticated_client.delete_project(p["id"])
    except APIError:
        pass


@pytest.fixture()
def endpoint_for_jobs(authenticated_client: UIAutomationClient):
    """Create a temporary endpoint for job tests."""
    ep = authenticated_client.create_endpoint(
        name="Job Test Endpoint",
        url="https://example.com",
    )
    yield ep
    try:
        authenticated_client.delete_endpoint(ep["id"])
    except APIError:
        pass


class TestCreateJob:
    def test_create_ui_exploration_job(
        self,
        authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
        endpoint_for_jobs: dict,
    ):
        resp = authenticated_client.create_job(
            job_type="ui_exploration",
            config={
                "endpoint_id": endpoint_for_jobs["id"],
                "project_id": project_for_jobs["id"],
                "procedure_name": "Test Exploration",
            },
        )
        assert "id" in resp
        assert resp["type"] == "ui_exploration"
        assert resp["status"] == "created"
        assert resp["config"]["endpoint_id"] == endpoint_for_jobs["id"]
        assert resp["config"]["project_id"] == project_for_jobs["id"]

    def test_create_job_invalid_type(
        self,
        authenticated_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_job(
                job_type="invalid_type",
                config={},
            )
        assert exc_info.value.status_code == 400

    def test_create_job_missing_endpoint_id(
        self,
        authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_job(
                job_type="ui_exploration",
                config={"project_id": project_for_jobs["id"]},
            )
        assert exc_info.value.status_code == 400

    def test_create_job_missing_project_id(
        self,
        authenticated_client: UIAutomationClient,
        endpoint_for_jobs: dict,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_job(
                job_type="ui_exploration",
                config={"endpoint_id": endpoint_for_jobs["id"]},
            )
        assert exc_info.value.status_code == 400

    def test_create_job_unauthenticated(
        self,
        fresh_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            fresh_client.create_job(
                job_type="ui_exploration",
                config={},
            )
        assert exc_info.value.status_code == 401


class TestListJobs:
    def test_list_jobs(
        self,
        authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
        endpoint_for_jobs: dict,
    ):
        # Create a job first
        job = authenticated_client.create_job(
            job_type="ui_exploration",
            config={
                "endpoint_id": endpoint_for_jobs["id"],
                "project_id": project_for_jobs["id"],
            },
        )
        resp = authenticated_client.list_jobs()
        assert "items" in resp
        assert "total" in resp
        assert resp["total"] >= 1
        ids = [j["id"] for j in resp["items"]]
        assert job["id"] in ids

    def test_list_jobs_pagination(
        self,
        authenticated_client: UIAutomationClient,
    ):
        resp = authenticated_client.list_jobs(limit=1, offset=0)
        assert resp["limit"] == 1
        assert resp["offset"] == 0


class TestGetJob:
    def test_get_job(
        self,
        authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
        endpoint_for_jobs: dict,
    ):
        job = authenticated_client.create_job(
            job_type="ui_exploration",
            config={
                "endpoint_id": endpoint_for_jobs["id"],
                "project_id": project_for_jobs["id"],
            },
        )
        resp = authenticated_client.get_job(job["id"])
        assert resp["id"] == job["id"]
        assert resp["type"] == "ui_exploration"
        assert resp["status"] == "created"

    def test_get_job_not_found(
        self,
        authenticated_client: UIAutomationClient,
    ):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.get_job(
                "00000000-0000-0000-0000-000000000000",
            )
        assert exc_info.value.status_code == 404


class TestStopJob:
    def test_stop_non_running_job_returns_400(
        self,
        authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
        endpoint_for_jobs: dict,
    ):
        job = authenticated_client.create_job(
            job_type="ui_exploration",
            config={
                "endpoint_id": endpoint_for_jobs["id"],
                "project_id": project_for_jobs["id"],
            },
        )
        # Job is in "created" status, not "running"
        with pytest.raises(APIError) as exc_info:
            authenticated_client.stop_job(job["id"])
        assert exc_info.value.status_code == 400


class TestJobStatusTransition:
    def test_job_transitions_to_running_after_creation(
        self,
        authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
        endpoint_for_jobs: dict,
    ):
        """After creating a ui_exploration job, it should transition from
        'created' to 'running' (or 'failed' if agent dependencies are missing).
        This verifies the pipeline goroutine is triggered."""
        job = authenticated_client.create_job(
            job_type="ui_exploration",
            config={
                "endpoint_id": endpoint_for_jobs["id"],
                "project_id": project_for_jobs["id"],
                "procedure_name": "Status Transition Test",
            },
        )
        assert job["status"] == "created"

        # Wait briefly for the pipeline goroutine to pick up the job
        time.sleep(2)

        updated = authenticated_client.get_job(job["id"])
        # The job should have transitioned away from "created".
        # It will be "running" if agent deps are available, or "failed"
        # if the Python script / Bedrock credentials are missing.
        assert updated["status"] in ("running", "failed", "success")


class TestJobOwnership:
    def test_other_user_cannot_access_job(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
        project_for_jobs: dict,
        endpoint_for_jobs: dict,
    ):
        job = authenticated_client.create_job(
            job_type="ui_exploration",
            config={
                "endpoint_id": endpoint_for_jobs["id"],
                "project_id": project_for_jobs["id"],
            },
        )
        with pytest.raises(APIError) as exc_info:
            second_authenticated_client.get_job(job["id"])
        assert exc_info.value.status_code == 403
