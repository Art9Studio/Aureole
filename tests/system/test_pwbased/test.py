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


@pytest.mark.skip()
def test_register(app_url):
    r = requests.post(app_url + '/register', json={
        'email': 'john.doe@example.com',
        'password': 'qwerty'
    })
    assert r.status_code == 200
    assert isinstance(r.json()['user_id'], int)

    r = requests.post(app_url + '/register', json={
        'email': 'john.doe@example.com',
        'password': 'qwerty'
    })
    assert r.status_code == 400
    assert r.json()['message'] == 'user already exist'


@pytest.mark.skip()
def test_login(app_url, uuid):
    login_resp = requests.post(app_url + '/login', json={
        'email': 'john.doe@example.com',
        'password': 'qwerty'
    })
    assert login_resp.ok
    assert login_resp.headers['access'] is not None

    cookie = SimpleCookie()
    cookie.load(login_resp.headers['Set-Cookie'])
    assert cookie.get('refresh_token') is not None
    assert verify_jwt(login_resp.headers['access'], cookie.get('refresh_token').value,
                      app_url + '/gen-keys/jwk', ['ES256'], uuid, 'Aureole Server')

    refresh_resp = requests.post(app_url + '/refresh', cookies=login_resp.cookies)
    assert refresh_resp.ok
    assert refresh_resp.headers['access'] is not None
    assert verify_jwt(refresh_resp.headers['access'], cookie.get('refresh_token').value,
                      app_url + '/gen-keys/jwk', ['ES256'], uuid, 'Aureole Server')


@pytest.mark.skip()
def test_email_verification(app_url):
    r = requests.post(app_url + '/email-verify', json={'email': 'john.doe@example.com'})
    assert r.ok

    emails = requests.get('http://smtp:1080/api/emails')
    link = emails.json()[0]['text']
    r = requests.get(link)
    assert r.ok

    requests.delete('http://smtp:1080/api/emails')


@pytest.mark.skip()
def test_password_reset(app_url):
    r = requests.post(app_url + '/password/reset', json={'email': 'john.doe@example.com'})
    assert r.ok

    emails = requests.get('http://smtp:1080/api/emails')
    link = emails.json()[0]['text']
    r = requests.post(link, json={'password': '1234'})
    assert r.ok

    requests.delete('http://smtp:1080/api/emails')
