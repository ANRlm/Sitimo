FROM golang:1.26-bookworm AS base

RUN apt-get update && apt-get install -y --no-install-recommends \
    imagemagick \
    libheif-examples \
    fonts-noto-cjk \
    texlive-xetex \
    texlive-lang-chinese \
    texlive-latex-extra \
    texlive-fonts-recommended \
    texlive-pictures \
    postgresql-client \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh

EXPOSE 8080

CMD ["./entrypoint.sh"]
