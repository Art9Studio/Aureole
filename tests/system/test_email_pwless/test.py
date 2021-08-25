import json
import os
from pathlib import Path

import pytest
import requests

from http.cookies import SimpleCookie

GLOBAL_RES_DIR = os.path.join(Path(__file__).parent.parent.resolve(), 'resources')
TEST_NAME = Path(__file__).parent.resolve().name
BASE_URL = 'http://aureole:3000'


@pytest.fixture(scope='module')
def uuid():
    uuid_path = os.path.join(GLOBAL_RES_DIR, 'uuid.txt')
    with open(uuid_path, 'r') as f:
        uuids = json.loads(f.read())
        return uuids[TEST_NAME]


@pytest.fixture(scope='module')
def app_url(uuid):
    return BASE_URL + f'/{uuid}'


def test_register(app_url):
    r = requests.post(app_url + '/register', json={'email': 'john.doe@example.com'})
    assert r.ok

    r = requests.post(app_url + '/register', json={'email': 'john.doe@example.com'})
    assert r.status_code == 400
    assert r.json()['message'] == 'user already exist'

    requests.delete('http://smtp:1080/api/emails')


def test_get_magic_link(app_url, uuid):
    r = requests.post(app_url + '/login', json={'email': 'john.doe@example.com'})
    assert r.ok

    emails = requests.get('http://smtp:1080/api/emails')
    link = emails.json()[0]['text']
    r = requests.get(link)
    assert r.ok
    assert r.headers['Set-Cookie'] is not None

    cookie = SimpleCookie()
    cookie.load(r.headers['Set-Cookie'])
    assert cookie.get('session_token') is not None

    requests.delete('http://smtp:1080/api/emails')
