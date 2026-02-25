"""End-to-end integration test that mirrors the original test_integration.sh flow."""

import base64
import os
import tempfile
import uuid

import pytest

from client import (
    ASSET_IMAGE,
    STATUS_PASSED,
    STATUS_PENDING,
    STATUS_RUNNING,
    UIAutomationClient,
)

pytestmark = pytest.mark.flow

# Minimal valid 1x1 PNG
_PNG_1X1 = base64.b64decode(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
)
_PNG_MAGIC = b"\x89PNG\r\n\x1a\n"


def test_full_integration_flow(base_url: str):
    """Replicate the full flow from test_integration.sh with real assertions."""
    suffix = uuid.uuid4().hex[:8]
    client = UIAutomationClient(base_url)

    # ── Scenario 1: Basic Project Flow ──

    # 1. Register user
    user = client.register(
        email=f"flow-{suffix}@example.com",
        username=f"flow-{suffix}",
        password="password12345678",
    )
    assert "id" in user
    assert user["email"] == f"flow-{suffix}@example.com"

    # 2. Login
    login_resp = client.login(
        f"flow-{suffix}@example.com", "password12345678",
    )
    assert "id" in login_resp

    # 2b. Validate session
    me = client.me()
    assert me["email"] == f"flow-{suffix}@example.com"

    # 3. Create project
    project = client.create_project(
        name="Test Project",
        description="Integration test project",
    )
    project_id = project["id"]
    assert project["name"] == "Test Project"

    # 4. List projects
    projects = client.list_projects(limit=10)
    assert projects["total"] >= 1
    project_ids = [p["id"] for p in projects["items"]]
    assert project_id in project_ids

    # 5. Get project by ID
    fetched = client.get_project(project_id)
    assert fetched["id"] == project_id
    assert fetched["name"] == "Test Project"

    # 6. Update project
    updated = client.update_project(project_id, name="Updated Project Name")
    assert updated["name"] == "Updated Project Name"

    # ── Scenario 2: Test Procedure with Versioning ──

    # 1. Create test procedure (correct step format)
    procedure = client.create_procedure(
        project_id=project_id,
        name="Login Test Procedure",
        description="Test login functionality",
        steps=[
            {
                "name": "Navigate to login",
                "instructions": "Open https://example.com/login in browser",
                "image_paths": [],
            },
            {
                "name": "Enter username",
                "instructions": "Type testuser into the #username field",
                "image_paths": [],
            },
            {
                "name": "Enter password",
                "instructions": "Type password into the #password field",
                "image_paths": [],
            },
            {
                "name": "Click login",
                "instructions": "Click the #login-button and verify redirect",
                "image_paths": [],
            },
        ],
    )
    procedure_id = procedure["id"]
    assert procedure["name"] == "Login Test Procedure"
    assert procedure["version"] == 1

    # 2. Create test run v1
    run1 = client.create_run(procedure_id)
    run1_id = run1["id"]
    assert run1["status"] == STATUS_PENDING

    # 3. Start test run
    started = client.start_run(run1_id)
    assert started["status"] == STATUS_RUNNING
    assert started["started_at"] is not None

    # 4. Complete test run
    completed = client.complete_run(
        run1_id, status=STATUS_PASSED, notes="All tests passed",
    )
    assert completed["status"] == STATUS_PASSED
    assert completed["completed_at"] is not None

    # 5. Update test procedure (in-place)
    updated_proc = client.update_procedure(
        project_id,
        procedure_id,
        description="Updated description for login test",
    )
    assert updated_proc["description"] == "Updated description for login test"

    # 6. Create new version
    version2 = client.create_version(project_id, procedure_id)
    version2_id = version2["id"]
    assert version2["id"] != procedure_id
    assert version2["version"] == 2
    assert version2["is_latest"] is True

    # 7. Get version history
    history = client.get_version_history(project_id, procedure_id)
    assert isinstance(history, list)
    assert len(history) >= 2

    # 8. Create test run with new version
    run2 = client.create_run(version2_id)
    run2_id = run2["id"]
    assert run2["status"] == STATUS_PENDING

    # ── Scenario 3: Asset Management ──

    # Need to start the run to upload assets
    client.start_run(run2_id)

    # Create a temp image file
    fd, image_path = tempfile.mkstemp(suffix=".png")
    os.write(fd, _PNG_1X1)
    os.close(fd)

    try:
        # 1. Upload image asset
        asset = client.upload_asset(
            run_id=run2_id,
            file_path=image_path,
            asset_type=ASSET_IMAGE,
            description="Test screenshot",
        )
        asset_id = asset["id"]
        assert asset["asset_type"] == ASSET_IMAGE
        assert asset["file_size"] > 0

        # 2. List assets
        assets = client.list_assets(run2_id)
        assert isinstance(assets, list)
        assert len(assets) >= 1
        asset_ids = [a["id"] for a in assets]
        assert asset_id in asset_ids

        # 3. Download asset
        data = client.download_asset(run2_id, asset_id)
        assert data[:8] == _PNG_MAGIC
        assert len(data) > 0

        # 4. Delete asset
        del_resp = client.delete_asset(run2_id, asset_id)
        assert "message" in del_resp
    finally:
        os.unlink(image_path)

    # ── Scenario 4: List All Data ──

    # List test procedures
    procedures = client.list_procedures(project_id)
    assert procedures["total"] >= 1

    # List test runs
    runs = client.list_runs(procedure_id)
    assert runs["total"] >= 1

    # Get test run details
    run_detail = client.get_run(run1_id)
    assert run_detail["id"] == run1_id
    assert run_detail["status"] == STATUS_PASSED

    # ── Cleanup ──

    # Delete project (cascade deletes procedures and runs)
    delete_resp = client.delete_project(project_id)
    assert "message" in delete_resp

    # Logout
    client.logout()

    # Verify session is invalidated
    from client import APIError
    with pytest.raises(APIError) as exc_info:
        client.me()
    assert exc_info.value.status_code == 401
