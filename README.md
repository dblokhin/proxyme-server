# Proxyme SOCKS5 server

[![Go Report Card](https://goreportcard.com/badge/github.com/dblokhin/proxyme-server)](https://goreportcard.com/report/github.com/dblokhin/proxyme-server)
[![Docker Pulls](https://img.shields.io/docker/pulls/dblokhin/proxyme)](https://hub.docker.com/r/dblokhin/proxyme)
[![Docker Image Size](https://img.shields.io/docker/image-size/dblokhin/proxyme)](https://hub.docker.com/r/dblokhin/proxyme)

This is an efficient and lightweight implementation of a Socks5 Proxy. The proxy supports key features like CONNECT, BIND, and AUTH (both with and without username/password 
authentication, and GSSAPI SOCKS5 authentication flow).

## Project Status
This project is currently **active** and maintained. We aim to continually improve its performance and feature set. 
Feedback and contributions are greatly appreciated!

## Features
This project fully implements all the requirements outlined in the specifications of RFC 1928, RFC 1929, and RFC 1961,
with the exception of the UDP ASSOCIATE command and GSSAPI auth method, which may be implemented in the future.

- **CONNECT command**: Standard command for connecting to a destination server.
- **Custom CONNECT**: Allows creating customs tunnels to destination server.
- **BIND command**: Allows incoming connections on a specified IP and port.
- **AUTH support**:
    - No authentication (anonymous access)
    - Username/Password authentication 
    - plan: GSSAPI SOCKS5 protocol flow (rfc1961)

## Getting Started
### Environment Variables
The project supports the following environment variables to configure the proxy server:

- `PROXY_HOST`: The host IP or hostname the proxy will listen on. (Default: 0.0.0.0)
- `PROXY_PORT`: The port number the proxy will listen on. (Default: 1080)
- `PROXY_BIND_IP`: The IP address to use for BIND operations in the SOCKS5 protocol. This should be a public IP address that can accept incoming connections. (Default: disabled)
- `PROXY_NOAUTH`: If set to yes, true, or 1, allows unauthenticated access to the proxy. (Default: disabled)
- `PROXY_USERS`: A comma-separated list of username and password pairs for authentication (in the format user:pass,user2:pass2). If this is set, the proxy enables SOCKS5 username/password authentication.

At least one SOCKS5 auth method (noauth or username/password) must be specified.

### Docker Usage
You can pull the ready-to-use image from Docker Hub: [https://hub.docker.com/r/dblokhin/proxyme](https://hub.docker.com/r/dblokhin/proxyme).

You can also run the socks5 proxy within a Docker container manually.

1. **Build the Docker image:**
   ```bash
   docker build -t proxyme .
   ```

2. **Run the Docker container:**
   ```bash
   docker run -d \
    -e PROXY_HOST=0.0.0.0 \
    -e PROXY_PORT=1080 \
    -e PROXY_BIND_IP=203.0.113.4 \
    -e PROXY_NOAUTH=yes \
    -e PROXY_USERS="user1:pass1,user2:pass2" \
    -p 1080:1080 \
    proxyme
   ```

   ```bash
   curl --socks5 localhost:1080 -U user1:pass1 https://google.com
   ```

### Binary installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/dblokhin/proxyme-server
   cd proxyme
   ```

2. **Build the binary:**
   ```bash
   make build
   ```

3. **Run the proxy:**
   ```bash
   PROXY_PORT=1080 PROXY_NOAUTH=yes ./proxyme # starts proxy on 0.0.0.0
   ```

4. **Check the proxy:**
   ```bash
   curl --socks5 localhost:1080 https://google.com
   ```
   
## Contributing
We welcome contributions to enhance the functionality and performance of this Socks5 proxy. If you find any bugs or have feature requests, feel free to open an issue or submit a pull request.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.


## References & useful links
* [RFC 1928: SOCKS Protocol Version 5](http://www.ietf.org/rfc/rfc1928.txt)
* [RFC 1929: Username/Password Authentication for SOCKS V5](http://www.ietf.org/rfc/rfc1929.txt)
* [RFC 1961: GSS-API Authentication Method for SOCKS Version 5](http://www.ietf.org/rfc/rfc1961.txt)

---

We encourage the community to contribute, experiment, and utilize this project for both learning and production purposes. If you are looking for easy-to-use SOCKS5 proxy written in Go, you have come to the right place!

Happy coding!
