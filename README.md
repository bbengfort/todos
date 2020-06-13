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

Once the user logs in, they will be granted a JWT token.

