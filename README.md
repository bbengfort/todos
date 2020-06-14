# TODOs

**Simple TODO API for personal task tracking (written in Go)**

Building task/todo apps seems to be the "shooting hoops" of development -- good practice to test a lot of different components and technologies. In this case, I'm working with the following technologies:

- Golang net/http API server (JSON with gin)
- Vue.js frontend single page app
- Docker and docker-compose for image management
- CI/CD with Travis-CI

This development is meant to answer several questions which I will place in this README and answer throughout the course of the project.

## Getting Started

TODO: write this section.

## Authentication

I've implemented password authentication using argon2 derived keys. [Argon2](https://cryptobook.nakov.com/mac-and-key-derivation/argon2) is a modern ASIC- and GPU- resistent secure key derivation function that stores passwords as a cryptographic hash in the database instead of plain text. The algorithm adds memory, time, and computational complexity to prevent rainbow and brute force attacks on a list of passwords stored this way. To compare passwords, you derive the key for the password and compare it to the derived key in the database without every saving it as plain text.

The database representation of derived keys is as follows:

```
$argon2id$v=19$m=65536,t=1,p=2$syEoYrFtsGBwudEnzzqvgw==$YPMFYzCdtdC1HEnQrxZlAj/Jl7HWLdqxcKqf7W4Om9w=
```

This standard format stores information needed to compute a per-user derived key with a `$` delimiter. The first two fields are the algorithm (argon2i, argon2d, or argon2id) along with the version of the argon implementation. The third field contains parameters for the key derivation function. The fourth and fifth fields are the user-specific 16 byte salt and the 32 byte derived key, both base64 encoded.

### JWT

Once the user logs in, they will be granted a JWT access token and JWT refresh token. I've done a lot of reading about API authorization schemes and honestly there is a lot of stuff out there. Frankly I'd prefer something like [HAWK](https://chrisdecairos.ca/api-key-authentication-using-hawk/) to JWT, but it seems like this scheme hasn't been updated since 2015. The way I intend to use JWT is similar to the database access/refresh method described [here](https://www.cloudjourney.io/articles/security/jwt_in_golang-su/) and [here](https://www.sohamkamani.com/golang/2019-01-01-jwt-authentication/), though with Postgres instead of Redis as a backend (this is basically a single user app).

So here's the scheme:

1. **Login**: grant an access token with a 4 hour expiration and a refresh token with a 12 hour expiration. These tokens are saved in the database with claims information. The login can optionally set a cookie.
2. **Authorization**: check the Bearer header, cookie, and token request parameters for the access token, verify the key still exists in the database and that it hasn't expired. Load user information into the context of the request or return unauthorized/anonymous.
3. **Logout**: fetch the access token in the same manner as Authorization, but then delete the token from the database, revoking access. The logout request can optionally take a "revoke all" parameter, which revokes all tokens for the user.
4. **Refresh**: the refresh token cannot be used for authorization, but it does have a longer expiration than the access token, which means that it can be used to periodically refresh the access token with clients (particularly the CLI clients).

Other features/notes:

- a new token is generated on every login, so the user can have different tokens on multiple devices.
- a side go routine needs to run periodically to clean up expired tokens or an automatic mechanism needs to delete the token from the database when it's expired.

