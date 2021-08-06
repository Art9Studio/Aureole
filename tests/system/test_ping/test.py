import json
import os
from pathlib import Path

import pytest
import requests

GLOBAL_RES_DIR = os.path.join(Path(__file__).parent.parent.resolve(), 'resources')
TEST_NAME = Path(__file__).parent.resolve().name
BASE_URL = 'http://aureole:3000'


@pytest.fixture(scope='module')
def uuid():
    uuid_path = os.path.join(GLOBAL_RES_DIR, 'uuid.txt')
    with open(uuid_path, 'r') as f:
        uuids = json.loads(f.read())
        return uuids[TEST_NAME]


def test_ping():
    assert requests.get(BASE_URL + '/ping').ok
