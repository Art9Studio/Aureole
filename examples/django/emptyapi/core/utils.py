import json

import jwt
import jwt.algorithms


def jwt_decode_token(token):
    header = jwt.get_unverified_header(token)

    with open('/examples/Django/emptyapi/core/keys.json') as jwk_file:
        jwk_set = json.load(jwk_file)

    public_key = None
    for jwk in jwk_set['keys']:
        if jwk['kid'] == header['kid']:
            public_key = jwt.algorithms.RSAAlgorithm.from_jwk(json.dumps(jwk))

    if public_key is None:
        raise Exception('Public key not found.')

    return jwt.decode(token, public_key, algorithms=['RS256'], audience=['emptyapi'], issuer='Aureole Server')
