import pytest

from client import (
    ASSET_IMAGE,
    APIError,
    UIAutomationClient,
)

pytestmark = pytest.mark.assets

SAMPLE_STEPS = [
    {
        "name": "Step 1",
        "instructions": "Take a screenshot",
        "image_paths": [],
    },
]

# PNG magic bytes
_PNG_MAGIC = b"\x89PNG\r\n\x1a\n"


@pytest.fixture()
def run_id(authenticated_client: UIAutomationClient):
    """Create project → procedure → run → start, yield run ID."""
    project = authenticated_client.create_project(
        name="Asset Test Project",
        description="For asset integration tests",
    )
    procedure = authenticated_client.create_procedure(
        project_id=project["id"],
        name="Asset Test Procedure",
        description="Procedure for asset tests",
        steps=SAMPLE_STEPS,
    )
    run = authenticated_client.create_run(procedure["id"])
    authenticated_client.start_run(run["id"])
    yield run["id"]
    try:
        authenticated_client.delete_project(project["id"])
    except APIError:
        pass


class TestUploadAsset:
    def test_upload_image(
        self,
        authenticated_client: UIAutomationClient,
        run_id: str,
        test_image_path: str,
    ):
        resp = authenticated_client.upload_asset(
            run_id=run_id,
            file_path=test_image_path,
            asset_type=ASSET_IMAGE,
            description="Test screenshot",
        )
        assert "id" in resp
        assert resp["asset_type"] == ASSET_IMAGE
        assert resp["description"] == "Test screenshot"
        assert resp["file_size"] > 0


class TestListAssets:
    def test_list_assets(
        self,
        authenticated_client: UIAutomationClient,
        run_id: str,
        test_image_path: str,
    ):
        authenticated_client.upload_asset(
            run_id=run_id,
            file_path=test_image_path,
            asset_type=ASSET_IMAGE,
            description="Listed asset",
        )
        assets = authenticated_client.list_assets(run_id)
        assert isinstance(assets, list)
        assert len(assets) >= 1
        assert assets[0]["asset_type"] == ASSET_IMAGE


class TestDownloadAsset:
    def test_download_returns_png(
        self,
        authenticated_client: UIAutomationClient,
        run_id: str,
        test_image_path: str,
    ):
        asset = authenticated_client.upload_asset(
            run_id=run_id,
            file_path=test_image_path,
            asset_type=ASSET_IMAGE,
            description="Download test",
        )
        data = authenticated_client.download_asset(run_id, asset["id"])
        assert data[:8] == _PNG_MAGIC


class TestDeleteAsset:
    def test_delete_asset(
        self,
        authenticated_client: UIAutomationClient,
        run_id: str,
        test_image_path: str,
    ):
        asset = authenticated_client.upload_asset(
            run_id=run_id,
            file_path=test_image_path,
            asset_type=ASSET_IMAGE,
            description="Will be deleted",
        )
        resp = authenticated_client.delete_asset(run_id, asset["id"])
        assert "message" in resp
