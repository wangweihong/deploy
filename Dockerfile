FROM busybox:latest
RUN mkdir /deploy


COPY ./ufleet-deploy /deploy
COPY ./conf /deploy/conf
COPY ./swagger /deploy/swagger

# ENV MODULE_VERSION #MODULE_VERSION#

WORKDIR /deploy
CMD ["/deploy/ufleet-deploy"]
