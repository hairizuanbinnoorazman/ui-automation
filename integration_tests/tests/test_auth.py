import uuid

import pytest

from client import APIError, UIAutomationClient

pytestmark = pytest.mark.auth


class TestRegister:
    def test_register_new_user(self, fresh_client: UIAutomationClient):
        suffix = uuid.uuid4().hex[:8]
        resp = fresh_client.register(
            email=f"reg-{suffix}@example.com",
            username=f"reg-{suffix}",
            password="password12345678",
        )
        assert "id" in resp
        assert resp["email"] == f"reg-{suffix}@example.com"
        assert resp["username"] == f"reg-{suffix}"

    def test_register_duplicate_email(
        self,
        fresh_client: UIAutomationClient,
        test_user_credentials: dict,
        authenticated_client,  # ensure user exists
    ):
        with pytest.raises(APIError) as exc_info:
            fresh_client.register(**test_user_credentials)
        assert exc_info.value.status_code == 409


class TestLogin:
    def test_login_success(
        self,
        fresh_client: UIAutomationClient,
        test_user_credentials: dict,
        authenticated_client,  # ensure user exists
    ):
        resp = fresh_client.login(
            test_user_credentials["email"],
            test_user_credentials["password"],
        )
        assert "id" in resp
        assert resp["email"] == test_user_credentials["email"]

    def test_login_wrong_password(
        self,
        fresh_client: UIAutomationClient,
        test_user_credentials: dict,
        authenticated_client,  # ensure user exists
    ):
        with pytest.raises(APIError) as exc_info:
            fresh_client.login(
                test_user_credentials["email"],
                "wrongpassword12345",
            )
        assert exc_info.value.status_code == 401


class TestMe:
    def test_me_authenticated(
        self,
        authenticated_client: UIAutomationClient,
        test_user_credentials: dict,
    ):
        resp = authenticated_client.me()
        assert resp["email"] == test_user_credentials["email"]
        assert resp["username"] == test_user_credentials["username"]

    def test_me_unauthenticated(self, fresh_client: UIAutomationClient):
        with pytest.raises(APIError) as exc_info:
            fresh_client.me()
        assert exc_info.value.status_code == 401


class TestLogout:
    def test_logout_invalidates_session(self, base_url: str):
        suffix = uuid.uuid4().hex[:8]
        client = UIAutomationClient(base_url)
        client.register(
            email=f"logout-{suffix}@example.com",
            username=f"logout-{suffix}",
            password="password12345678",
        )
        client.login(f"logout-{suffix}@example.com", "password12345678")

        # Session is valid
        client.me()

        # Logout
        client.logout()

        # Session should be invalid
        with pytest.raises(APIError) as exc_info:
            client.me()
        assert exc_info.value.status_code == 401
