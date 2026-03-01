import os
from dataclasses import dataclass

import requests


@dataclass
class APIError(Exception):
    status_code: int
    body: dict | str

    def __str__(self) -> str:
        return f"APIError({self.status_code}): {self.body}"


class UIAutomationClient:
    """HTTP client for the UI Automation backend API."""

    def __init__(self, base_url: str | None = None) -> None:
        if base_url is None:
            base_url = os.environ.get("BASE_URL")
        if base_url is None:
            port = os.environ.get("APP_PORT", "8080")
            base_url = f"http://localhost:{port}"

        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()

    def _url(self, path: str) -> str:
        return f"{self.base_url}/api/v1{path}"

    def _request(self, method: str, path: str, **kwargs) -> dict | list:
        resp = self.session.request(method, self._url(path), **kwargs)
        if not resp.ok:
            try:
                body = resp.json()
            except (ValueError, requests.exceptions.JSONDecodeError):
                body = resp.text
            raise APIError(status_code=resp.status_code, body=body)
        if resp.status_code == 204 or not resp.content:
            return {}
        return resp.json()

    def _raw_request(self, method: str, path: str, **kwargs) -> requests.Response:
        resp = self.session.request(method, self._url(path), **kwargs)
        if not resp.ok:
            try:
                body = resp.json()
            except (ValueError, requests.exceptions.JSONDecodeError):
                body = resp.text
            raise APIError(status_code=resp.status_code, body=body)
        return resp

    # --- Auth ---

    def register(self, email: str, username: str, password: str) -> dict:
        return self._request("POST", "/auth/register", json={
            "email": email,
            "username": username,
            "password": password,
        })

    def login(self, email: str, password: str) -> dict:
        return self._request("POST", "/auth/login", json={
            "email": email,
            "password": password,
        })

    def me(self) -> dict:
        return self._request("GET", "/auth/me")

    def logout(self) -> dict:
        return self._request("POST", "/auth/logout")

    # --- Projects ---

    def create_project(self, name: str, description: str = "") -> dict:
        return self._request("POST", "/projects", json={
            "name": name,
            "description": description,
        })

    def list_projects(self, limit: int = 20, offset: int = 0) -> dict:
        return self._request("GET", "/projects", params={
            "limit": limit,
            "offset": offset,
        })

    def get_project(self, project_id: str) -> dict:
        return self._request("GET", f"/projects/{project_id}")

    def update_project(self, project_id: str, **fields) -> dict:
        return self._request("PUT", f"/projects/{project_id}", json=fields)

    def delete_project(self, project_id: str) -> dict:
        return self._request("DELETE", f"/projects/{project_id}")

    # --- Test Procedures ---

    def create_procedure(
        self,
        project_id: str,
        name: str,
        description: str = "",
        steps: list[dict] | None = None,
    ) -> dict:
        payload: dict = {"name": name, "description": description}
        if steps is not None:
            payload["steps"] = steps
        return self._request(
            "POST", f"/projects/{project_id}/procedures", json=payload,
        )

    def list_procedures(
        self, project_id: str, limit: int = 20, offset: int = 0,
    ) -> dict:
        return self._request(
            "GET", f"/projects/{project_id}/procedures",
            params={"limit": limit, "offset": offset},
        )

    def get_procedure(self, project_id: str, procedure_id: str) -> dict:
        return self._request(
            "GET", f"/projects/{project_id}/procedures/{procedure_id}",
        )

    def update_procedure(
        self, project_id: str, procedure_id: str, **fields,
    ) -> dict:
        return self._request(
            "PUT", f"/projects/{project_id}/procedures/{procedure_id}",
            json=fields,
        )

    def create_version(self, project_id: str, procedure_id: str) -> dict:
        return self._request(
            "POST",
            f"/projects/{project_id}/procedures/{procedure_id}/versions",
        )

    def get_version_history(
        self, project_id: str, procedure_id: str,
    ) -> list:
        return self._request(
            "GET",
            f"/projects/{project_id}/procedures/{procedure_id}/versions",
        )

    # --- Test Runs ---

    def create_run(self, procedure_id: str) -> dict:
        return self._request("POST", f"/procedures/{procedure_id}/runs")

    def list_runs(
        self, procedure_id: str, limit: int = 20, offset: int = 0,
    ) -> dict:
        return self._request(
            "GET", f"/procedures/{procedure_id}/runs",
            params={"limit": limit, "offset": offset},
        )

    def get_run(self, run_id: str) -> dict:
        return self._request("GET", f"/runs/{run_id}")

    def update_run(self, run_id: str, **fields) -> dict:
        return self._request("PUT", f"/runs/{run_id}", json=fields)

    def start_run(self, run_id: str) -> dict:
        return self._request("POST", f"/runs/{run_id}/start")

    def complete_run(self, run_id: str, status: str, notes: str = "") -> dict:
        payload: dict = {"status": status}
        if notes:
            payload["notes"] = notes
        return self._request("POST", f"/runs/{run_id}/complete", json=payload)

    # --- Assets ---

    def upload_asset(
        self,
        run_id: str,
        file_path: str,
        asset_type: str,
        description: str = "",
        step_index: int | None = None,
    ) -> dict:
        with open(file_path, "rb") as f:
            files = {"file": (os.path.basename(file_path), f)}
            data: dict = {"asset_type": asset_type}
            if description:
                data["description"] = description
            if step_index is not None:
                data["step_index"] = str(step_index)
            return self._request(
                "POST", f"/runs/{run_id}/assets", files=files, data=data,
            )

    def list_assets(self, run_id: str) -> list:
        return self._request("GET", f"/runs/{run_id}/assets")

    def download_asset(self, run_id: str, asset_id: str) -> bytes:
        resp = self._raw_request("GET", f"/runs/{run_id}/assets/{asset_id}")
        return resp.content

    def delete_asset(self, run_id: str, asset_id: str) -> dict:
        return self._request("DELETE", f"/runs/{run_id}/assets/{asset_id}")

    # --- Users ---

    def list_users(self, search: str = "", limit: int = 20, offset: int = 0) -> dict:
        params: dict = {"limit": limit, "offset": offset}
        if search:
            params["search"] = search
        return self._request("GET", "/users", params=params)

    def assign_run(self, run_id: str, user_id: str) -> dict:
        return self._request("PUT", f"/runs/{run_id}", json={"assigned_to": user_id})

    def unassign_run(self, run_id: str) -> dict:
        return self._request("PUT", f"/runs/{run_id}", json={"assigned_to": ""})

    # --- Endpoints ---

    def create_endpoint(
        self,
        name: str,
        url: str,
        credentials: list[dict] | None = None,
    ) -> dict:
        payload: dict = {"name": name, "url": url}
        if credentials is not None:
            payload["credentials"] = credentials
        return self._request("POST", "/endpoints", json=payload)

    def list_endpoints(self, limit: int = 20, offset: int = 0) -> dict:
        return self._request(
            "GET", "/endpoints",
            params={"limit": limit, "offset": offset},
        )

    def get_endpoint(self, endpoint_id: str) -> dict:
        return self._request("GET", f"/endpoints/{endpoint_id}")

    def update_endpoint(self, endpoint_id: str, **fields) -> dict:
        return self._request(
            "PUT", f"/endpoints/{endpoint_id}", json=fields,
        )

    def delete_endpoint(self, endpoint_id: str) -> dict:
        return self._request("DELETE", f"/endpoints/{endpoint_id}")

    # --- Jobs ---

    def create_job(
        self,
        job_type: str,
        config: dict | None = None,
    ) -> dict:
        payload: dict = {"type": job_type}
        if config is not None:
            payload["config"] = config
        return self._request("POST", "/jobs", json=payload)

    def list_jobs(self, limit: int = 20, offset: int = 0) -> dict:
        return self._request(
            "GET", "/jobs",
            params={"limit": limit, "offset": offset},
        )

    def get_job(self, job_id: str) -> dict:
        return self._request("GET", f"/jobs/{job_id}")

    def stop_job(self, job_id: str) -> dict:
        return self._request("POST", f"/jobs/{job_id}/stop")

    # --- API Tokens ---

    def create_api_token(
        self,
        name: str,
        scope: str = "read_only",
        expires_in_hours: int = 720,
    ) -> dict:
        return self._request("POST", "/tokens", json={
            "name": name,
            "scope": scope,
            "expires_in_hours": expires_in_hours,
        })

    def list_api_tokens(self) -> dict:
        return self._request("GET", "/tokens")

    def revoke_api_token(self, token_id: str) -> dict:
        return self._request("DELETE", f"/tokens/{token_id}")

    # --- Integrations ---

    def create_integration(
        self,
        name: str,
        provider: str,
        credentials: list[dict],
    ) -> dict:
        return self._request("POST", "/integrations", json={
            "name": name,
            "provider": provider,
            "credentials": credentials,
        })

    def list_integrations(self) -> dict:
        return self._request("GET", "/integrations")

    def get_integration(self, integration_id: str) -> dict:
        return self._request("GET", f"/integrations/{integration_id}")

    def update_integration(self, integration_id: str, **fields) -> dict:
        return self._request(
            "PUT", f"/integrations/{integration_id}", json=fields,
        )

    def delete_integration(self, integration_id: str) -> dict:
        return self._request("DELETE", f"/integrations/{integration_id}")

    def test_integration_connection(self, integration_id: str) -> dict:
        return self._request("POST", f"/integrations/{integration_id}/test")

    def search_external_issues(
        self,
        integration_id: str,
        query: str = "",
    ) -> dict:
        params = {}
        if query:
            params["query"] = query
        return self._request(
            "GET", f"/integrations/{integration_id}/issues",
            params=params,
        )

    # --- Issue Links ---

    def create_and_link_issue(
        self,
        run_id: str,
        integration_id: str,
        title: str,
        description: str = "",
        project_key: str = "",
        issue_type: str = "",
        repository: str = "",
        labels: list[str] | None = None,
    ) -> dict:
        payload: dict = {
            "integration_id": integration_id,
            "title": title,
            "description": description,
        }
        if project_key:
            payload["project_key"] = project_key
        if issue_type:
            payload["issue_type"] = issue_type
        if repository:
            payload["repository"] = repository
        if labels:
            payload["labels"] = labels
        return self._request(
            "POST", f"/runs/{run_id}/issues", json=payload,
        )

    def link_existing_issue(
        self,
        run_id: str,
        integration_id: str,
        external_id: str,
    ) -> dict:
        return self._request(
            "POST", f"/runs/{run_id}/issues/link", json={
                "integration_id": integration_id,
                "external_id": external_id,
            },
        )

    def list_issue_links(self, run_id: str) -> list:
        return self._request("GET", f"/runs/{run_id}/issues")

    def unlink_issue(self, run_id: str, link_id: str) -> dict:
        return self._request(
            "DELETE", f"/runs/{run_id}/issues/{link_id}",
        )

    def resolve_linked_issue(self, run_id: str, link_id: str) -> dict:
        return self._request(
            "POST", f"/runs/{run_id}/issues/{link_id}/resolve",
        )

    def sync_issue_status(self, run_id: str, link_id: str) -> dict:
        return self._request(
            "POST", f"/runs/{run_id}/issues/{link_id}/sync",
        )

    def request_with_token(self, method: str, path: str, token: str, **kwargs) -> dict:
        """Make an API request using a Bearer token instead of session cookies."""
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.request(
            method, self._url(path), headers=headers, **kwargs,
        )
        if not resp.ok:
            try:
                body = resp.json()
            except (ValueError, requests.exceptions.JSONDecodeError):
                body = resp.text
            raise APIError(status_code=resp.status_code, body=body)
        if resp.status_code == 204 or not resp.content:
            return {}
        return resp.json()
