import json
import os
from http.cookies import SimpleCookie
from pathlib import Path

import pytest
import requests

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


def test_google_login(app_url, uuid):
    r = requests.get(app_url + '/oauth2/google', verify=False)
    assert r.ok
    assert r.headers['access'] is not None

    cookie = SimpleCookie()
    cookie.load(r.headers['Set-Cookie'])
    assert cookie.get('refresh_token') is not None
    assert verify_jwt(r.headers['access'], cookie.get('refresh_token').value,
                      app_url + '/keys/jwk', ['RS256'], uuid, 'Aureole Server')


def test_apple_login(app_url, uuid):
    r = requests.get(app_url + '/oauth2/apple', verify=False)
    assert r.ok
    assert r.headers['access'] is not None

    cookie = SimpleCookie()
    cookie.load(r.headers['Set-Cookie'])
    assert cookie.get('refresh_token') is not None
    assert verify_jwt(r.headers['access'], cookie.get('refresh_token').value,
                      app_url + '/keys/jwk', ['RS256'], uuid, 'Aureole Server')


def test_vk_login(app_url, uuid):
    r = requests.get(app_url + '/oauth2/vk', verify=False)
    assert r.ok
    assert r.headers['access'] is not None

    cookie = SimpleCookie()
    cookie.load(r.headers['Set-Cookie'])
    assert cookie.get('refresh_token') is not None
    assert verify_jwt(r.headers['access'], cookie.get('refresh_token').value,
                      app_url + '/keys/jwk', ['RS256'], uuid, 'Aureole Server')


def test_facebook_login(app_url, uuid):
    r = requests.get(app_url + '/oauth2/facebook', verify=False)
    assert r.ok
    assert r.headers['access'] is not None

    cookie = SimpleCookie()
    cookie.load(r.headers['Set-Cookie'])
    assert cookie.get('refresh_token') is not None
    assert verify_jwt(r.headers['access'], cookie.get('refresh_token').value,
                      app_url + '/keys/jwk', ['RS256'], uuid, 'Aureole Server')
