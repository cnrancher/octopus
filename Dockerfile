FROM scratch
WORKDIR /
VOLUME /var/lib/octopus/adaptors
COPY dist/octopus /
ENTRYPOINT ["/octopus"]
