FROM golang
RUN wget https://musl.libc.org/releases/musl-1.2.5.tar.gz && \
   tar -xzf musl-1.2.5.tar.gz && \
   cd musl-1.2.5 && \
   ./configure --enable-static --disable-shared && \
   make && make install

WORKDIR /src

ENV PATH="/usr/local/musl/bin:$PATH"
ENV CGO_ENABLED=1
ENV CC=musl-gcc

CMD go build --ldflags '-linkmode external -extldflags=-static' -tags "json1 fts5" -o pod-babashka-go-sqlite3 main.go
