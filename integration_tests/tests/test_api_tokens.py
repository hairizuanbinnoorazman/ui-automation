import pytest

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.tokens


class TestCreateAPIToken:
    def test_create_default(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_api_token(name="test-default")
        assert "id" in resp
        assert resp["name"] == "test-default"
        assert resp["scope"] == "read_only"
        assert "token" in resp
        assert resp["token"].startswith("uat_")
        assert "expires_at" in resp
        assert "created_at" in resp
        # Cleanup
        authenticated_client.revoke_api_token(resp["id"])

    def test_create_read_write(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_api_token(
            name="rw-token", scope="read_write",
        )
        assert resp["scope"] == "read_write"
        authenticated_client.revoke_api_token(resp["id"])

    def test_create_missing_name(self, authenticated_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_api_token(name="")
        assert exc_info.value.status_code == 400

    def test_create_invalid_scope(self, authenticated_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.create_api_token(name="bad-scope", scope="admin")
        assert exc_info.value.status_code == 400

    def test_create_unauthenticated(self, fresh_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            fresh_client.create_api_token(name="no-auth")
        assert exc_info.value.status_code == 401

    def test_create_max_limit(self, authenticated_client: UIAutomationClient):
        """Creating more than 5 active tokens should fail with 409."""
        token_ids = []
        try:
            for i in range(5):
                resp = authenticated_client.create_api_token(
                    name=f"limit-token-{i}",
                )
                token_ids.append(resp["id"])

            with pytest.raises(APIError) as exc_info:
                authenticated_client.create_api_token(name="one-too-many")
            assert exc_info.value.status_code == 409
        finally:
            for tid in token_ids:
                try:
                    authenticated_client.revoke_api_token(tid)
                except APIError:
                    pass


class TestListAPITokens:
    def test_list_shape(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.list_api_tokens()
        assert "tokens" in resp
        assert "total" in resp
        assert isinstance(resp["tokens"], list)

    def test_no_secret_in_list(self, authenticated_client: UIAutomationClient):
        create_resp = authenticated_client.create_api_token(name="list-check")
        try:
            list_resp = authenticated_client.list_api_tokens()
            for token in list_resp["tokens"]:
                assert "token" not in token, "token secret should not be in list"
                assert "id" in token
                assert "name" in token
                assert "scope" in token
                assert "expires_at" in token
                assert "is_active" in token
                assert "created_at" in token
        finally:
            authenticated_client.revoke_api_token(create_resp["id"])


class TestRevokeAPIToken:
    def test_revoke_success(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_api_token(name="to-revoke")
        authenticated_client.revoke_api_token(resp["id"])

        # Verify token no longer in active list
        list_resp = authenticated_client.list_api_tokens()
        active_ids = [t["id"] for t in list_resp["tokens"]]
        assert resp["id"] not in active_ids

    def test_revoke_other_user_forbidden(
        self,
        authenticated_client: UIAutomationClient,
        second_authenticated_client: UIAutomationClient,
    ):
        resp = authenticated_client.create_api_token(name="owned-token")
        try:
            with pytest.raises(APIError) as exc_info:
                second_authenticated_client.revoke_api_token(resp["id"])
            assert exc_info.value.status_code == 403
        finally:
            authenticated_client.revoke_api_token(resp["id"])


class TestBearerTokenAuth:
    def test_bearer_auth_works(self, authenticated_client: UIAutomationClient):
        """A valid Bearer token should authenticate API requests."""
        resp = authenticated_client.create_api_token(
            name="bearer-test", scope="read_write",
        )
        raw_token = resp["token"]
        try:
            # Use the token to hit a read endpoint
            result = authenticated_client.request_with_token(
                "GET", "/projects", raw_token, params={"limit": 1},
            )
            assert "items" in result
        finally:
            authenticated_client.revoke_api_token(resp["id"])

    def test_invalid_token_rejected(self, authenticated_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            authenticated_client.request_with_token(
                "GET", "/projects", "uat_invalid_token_value",
            )
        assert exc_info.value.status_code == 401

    def test_revoked_token_rejected(self, authenticated_client: UIAutomationClient):
        resp = authenticated_client.create_api_token(name="revoke-then-use")
        raw_token = resp["token"]
        authenticated_client.revoke_api_token(resp["id"])

        with pytest.raises(APIError) as exc_info:
            authenticated_client.request_with_token(
                "GET", "/projects", raw_token,
            )
        assert exc_info.value.status_code == 401
