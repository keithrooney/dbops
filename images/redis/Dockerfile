ARG UBUNTU_VERSION=20.04

FROM ubuntu:${UBUNTU_VERSION} as build

RUN apt-get upgrade -y
RUN apt-get update -y

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get install gcc wget make pkg-config -y

ARG REDIS_VERSION=7.0.2
ARG REDIS_DIRECTORY=/opt/redis

RUN mkdir -p ${REDIS_DIRECTORY}/downloads

RUN wget https://github.com/redis/redis/archive/${REDIS_VERSION}.tar.gz -O ${REDIS_DIRECTORY}/downloads/redis-${REDIS_VERSION}.tar.gz
RUN tar -xf ${REDIS_DIRECTORY}/downloads/redis-${REDIS_VERSION}.tar.gz -C ${REDIS_DIRECTORY}
RUN ln -s ${REDIS_DIRECTORY}/redis-${REDIS_VERSION} ${REDIS_DIRECTORY}/latest

RUN cd ${REDIS_DIRECTORY}/latest && make install

FROM ubuntu:${UBUNTU_VERSION}

RUN apt-get upgrade -y
RUN apt-get update -y

COPY --from=build /usr/local/bin/redis-server /usr/local/bin/
COPY --from=build /usr/local/bin/redis-sentinel /usr/local/bin/

EXPOSE 6379
