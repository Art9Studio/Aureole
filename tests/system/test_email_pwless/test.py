import json
import os
from pathlib import Path

import pytest
import requests

from http.cookies import SimpleCookie
from system.jwt_verify.jwt import verify_jwt

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


@pytest.mark.skip()
def test_register(app_url):
    r = requests.post(app_url + '/register', json={'email': 'john.doe@example.com'})
    assert r.ok

    r = requests.post(app_url + '/register', json={'email': 'john.doe@example.com'})
    assert r.status_code == 400
    assert r.json()['message'] == 'user already exist'

    requests.delete('http://smtp:1080/api/emails')


def test_login(app_url, uuid):
    r = requests.post(app_url + '/email-link/send', json={'email': 'john.doe@example.com'})
    assert r.ok

    emails = requests.get('http://smtp:1080/api/emails')
    link = emails.json()[0]['text']
    r = requests.get(link)
    assert r.ok
    assert r.headers['access'] is not None
    assert r.headers['Set-Cookie'] is not None

    cookie = SimpleCookie()
    cookie.load(r.headers['Set-Cookie'])
    assert cookie.get('refresh_token') is not None
    assert verify_jwt(r.headers['access'], cookie.get('refresh_token').value,
                      BASE_URL + '/email-pwless-jwk-file/jwk', ['RS256'], uuid, 'Aureole Server')

    requests.delete('http://smtp:1080/api/emails')
