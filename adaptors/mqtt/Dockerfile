FROM --platform=$TARGETPLATFORM scratch

# NB(thxCode): automatic platform ARGs, ref to:
# - https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /
VOLUME /var/lib/octopus/adaptors
COPY bin/mqtt_${TARGETOS}_${TARGETARCH} /mqtt
ENTRYPOINT ["/mqtt"]
