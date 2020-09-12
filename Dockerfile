FROM openjdk:13-jdk-alpine as jdk
RUN apk add --no-cache git

FROM golang:1.15-alpine as spigot-cli-builder
WORKDIR /spigot-cli
COPY . .
RUN apk add --update git
RUN go install

FROM jdk as spigot-jdk-builder
ENV REV=1.16.2
ADD https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar ./BuildTools.jar
RUN java -jar BuildTools.jar --rev ${REV} && mv spigot-${REV}.jar spigot.jar

FROM jdk
COPY --from=spigot-cli-builder ${HOME}/go/bin/spigot-cli /usr/local/bin/spigot-cli
COPY --from=spigot-jdk-builder spigot.jar spigot.jar
WORKDIR /server
ENTRYPOINT spigot-cli --spigot-path="../spigot.jar" start

