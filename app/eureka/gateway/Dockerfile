ARG MAVEN_VERSION=3.9.5
ARG JDK_VERSION=17.0.8
ARG IMAGE_PREFIX=mvkvl
ARG IMG_SUFFIX
ARG JAR_FILE="gateway-0.0.1.jar"
ARG MVNCMD="mvn -B -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn --no-transfer-progress --quiet -DskipTests -Dmaven.test.skip -am -T 4"

# -----------------------------------------------------------------------------------------------
FROM maven:3.9.6-sapmachine-21 AS prepare
ARG JDK_VERSION
ARG IMG_SUFFIX
ARG JAR_FILE
ARG VERSION
ARG MODULE
ARG MVNCMD

COPY   ./    /src
WORKDIR      /src

RUN apt-get update && apt install -y zip binutils

FROM prepare AS build
ARG JDK_VERSION
ARG IMG_SUFFIX
ARG JAR_FILE
ARG VERSION
ARG MODULE
ARG MVNCMD

RUN ${MVNCMD} -Dmaven.repo.local=/src/m2 -pl "${MODULE}" clean package

FROM build AS package
ARG JDK_VERSION
ARG IMG_SUFFIX
ARG JAR_FILE
ARG VERSION
ARG MODULE
ARG MVNCMD

# extract JAR Layers
WORKDIR /src/target
RUN java -Djarmode=layertools -jar ${JAR_FILE} list
RUN java -Djarmode=layertools -jar ${JAR_FILE} extract

# create slim JRE
RUN unzip ${JAR_FILE} -d temp > /dev/null
RUN jdeps                                 \
    --print-module-deps                   \
    --ignore-missing-deps                 \
    --recursive                           \
    --multi-release 17                    \
    --class-path="./temp/BOOT-INF/lib/*"  \
    --module-path="./temp/BOOT-INF/lib/*" \
    ${JAR_FILE} > /modules.txt
RUN jlink                                               \
    --add-modules $(cat /modules.txt | grep -v Warning) \
    --strip-debug                                       \
    --no-man-pages                                      \
    --no-header-files                                   \
    --compress=2                                        \
    --output /jre

ENTRYPOINT ["/bin/sh"]

# -----------------------------------------------------------------------------------------------
# debian-base
FROM debian:10-slim as base-debian
ARG IMG_SUFFIX

# -----------------------------------------------------------------------------------------------
## package
FROM base-debian as runtime
ARG MODULE
ARG BUILD_NUMBER
ENV JAVA_HOME=/jre
ENV PATH="${JAVA_HOME}/bin:${PATH}"
ENV BUILD_NUMBER=${BUILD_NUMBER}

# setup locales
RUN echo 'en_US.UTF-8 UTF-8' > /etc/locale.gen && apt-get update && apt-get -y install locales && locale-gen \
    && apt-get clean autoclean \
    && apt-get autoremove --yes \
    && rm -rf /var/lib/{apt,dpkg,cache,log}
ENV LANG=en_US.UTF-8

# add non-root user
RUN addgroup --system nonroot && adduser --system --ingroup nonroot nonroot
USER nonroot:nonroot

# copy slim JRE
COPY --from=package --chown=nonroot:nonroot /jre ${JAVA_HOME}

# copy layers from build image to runtime image as nonroot user
WORKDIR /application
COPY --from=package --chown=nonroot:nonroot /src/${MODULE}/target/dependencies/          ./
COPY --from=package --chown=nonroot:nonroot /src/${MODULE}/target/snapshot-dependencies/ ./
COPY --from=package --chown=nonroot:nonroot /src/${MODULE}/target/spring-boot-loader/    ./
COPY --from=package --chown=nonroot:nonroot /src/${MODULE}/target/application/           ./

EXPOSE 8080
ENTRYPOINT /jre/bin/java ${JAVA_OPTS} org.springframework.boot.loader.JarLauncher --spring.config.location=file:///application/BOOT-INF/classes/application.properties,file:///conf/routes.yml
