import base64
import os
import tempfile
import uuid

import pytest

from client import UIAutomationClient

# Minimal valid 1x1 red PNG (67 bytes)
_PNG_1X1 = base64.b64decode(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
)


@pytest.fixture(scope="session")
def base_url() -> str:
    url = os.environ.get("BASE_URL")
    if url:
        return url
    port = os.environ.get("APP_PORT", "8080")
    return f"http://localhost:{port}"


@pytest.fixture(scope="session")
def unique_suffix() -> str:
    return uuid.uuid4().hex[:8]


@pytest.fixture(scope="session")
def test_user_credentials(unique_suffix: str) -> dict:
    return {
        "email": f"test-{unique_suffix}@example.com",
        "username": f"testuser-{unique_suffix}",
        "password": "password12345678",
    }


@pytest.fixture(scope="session")
def authenticated_client(
    base_url: str, test_user_credentials: dict,
) -> UIAutomationClient:
    client = UIAutomationClient(base_url)
    client.register(**test_user_credentials)
    client.login(
        test_user_credentials["email"],
        test_user_credentials["password"],
    )
    yield client
    try:
        client.logout()
    except Exception:
        pass


@pytest.fixture(scope="session")
def second_user_credentials(unique_suffix: str) -> dict:
    return {
        "email": f"test2-{unique_suffix}@example.com",
        "username": f"testuser2-{unique_suffix}",
        "password": "password12345678",
    }


@pytest.fixture(scope="session")
def second_authenticated_client(
    base_url: str, second_user_credentials: dict,
) -> UIAutomationClient:
    client = UIAutomationClient(base_url)
    client.register(**second_user_credentials)
    client.login(
        second_user_credentials["email"],
        second_user_credentials["password"],
    )
    yield client
    try:
        client.logout()
    except Exception:
        pass


@pytest.fixture()
def fresh_client(base_url: str) -> UIAutomationClient:
    return UIAutomationClient(base_url)


@pytest.fixture()
def test_image_path() -> str:
    fd, path = tempfile.mkstemp(suffix=".png")
    os.write(fd, _PNG_1X1)
    os.close(fd)
    yield path
    try:
        os.unlink(path)
    except OSError:
        pass
