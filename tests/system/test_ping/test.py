import os
from pathlib import Path

import pytest
import requests

WORK_DIR = Path(__file__).parent.resolve()


@pytest.fixture(scope='module')
def uuid():
    uuid_path = os.path.join(WORK_DIR, './resources/uuid')
    with open(uuid_path, 'r') as f:
        return f.read()


def test_1(uuid):
    print(f'-  test_1({uuid})', __name__)


def test_ping():
    assert requests.get('http://localhost:3000/ping').status_code == 200