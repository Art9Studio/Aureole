import jwt
from jwt import PyJWKClient


def verify_jwt(access_token, refresh_token, jwk_url, alg, aud, iss):
    jwks_client = PyJWKClient(jwk_url)
    access_key = jwks_client.get_signing_key_from_jwt(access_token)
    refresh_key = jwks_client.get_signing_key_from_jwt(refresh_token)

    try:
        jwt.decode(access_token, access_key.key, algorithms=alg, audience=aud, issuer=iss)
        jwt.decode(refresh_token, refresh_key.key, algorithms=alg, issuer=iss)
    except jwt.PyJWTError:
        return False

    return True
