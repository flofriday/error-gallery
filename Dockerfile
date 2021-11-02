FROM ubuntu:latest

RUN apt update -y
RUN apt install build-essential gcc clang curl wget -y
RUN gcc --version && clang --version

# Install rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustc --version

# Install go
RUN rm -rf /usr/local/go && wget https://golang.org/dl/go1.17.2.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.17.2.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"
RUN go version

# Install elm
RUN curl -L -o elm.gz https://github.com/elm/compiler/releases/download/0.19.1/binary-for-linux-64-bit.gz && gunzip elm.gz && chmod +x elm && mv elm /usr/local/bin/
RUN elm --version

# Install the dependencies
COPY go.mod go.sum ./
RUN go mod download

WORKDIR /PWD
CMD ["go", "run", "generate.go"]
